package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = "ghedgetrim - %s\n"
)

// VERSION is set via ldflags & git tags
var VERSION string

var (
	token       string
	interval    string
	rawBranches string
	lastChecked time.Time

	branches = []string{"master"} // master is *always* protected

	debug   bool
	version bool
)

func init() {
	// parse flags
	flag.StringVar(&token, "token", "", "GitHub API token")
	flag.StringVar(&rawBranches, "branches", "master, develop", "protected branches, comma seperated)")
	flag.StringVar(&interval, "interval", "30s", "check interval (ex. 5ms, 10s, 1m, 3h)")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if token == "" {
		usageAndExit("GitHub token cannot be empty.", 1)
	}

	// convert rawBranches to an array
	for _, i := range strings.Split(rawBranches, ",") {
		branches = append(branches, strings.TrimSpace(i))
	}
	removeDuplicates(&branches)
	logrus.Infof("Protected branches %v.", branches)
}

func main() {
	var ticker *time.Ticker
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			ticker.Stop()
			logrus.Infof("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	// Create the http client.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	// Create the github client.
	client := github.NewClient(tc)

	// Get the authenticated user
	username := whoAmI(client)

	// parse the duration
	dur, err := time.ParseDuration(interval)
	if err != nil {
		logrus.Fatalf("parsing %s as duration failed: %v", interval, err)
	}
	ticker = time.NewTicker(dur)

	logrus.Infof("Bot started for user %s.", username)

	for range ticker.C {
		page := 1
		perPage := 20
		if err := getIssues(client, username, page, perPage); err != nil {
			logrus.Warn(err)
		}
	}
}

// Get the authenticated user, the empty string being passed let's the GitHub
// API know we want ourself. If we don't know who we are, we bail.
func whoAmI(client *github.Client) string {
	user, _, err := client.Users.Get("")
	if err != nil {
		logrus.Fatal(err)
	}
	return *user.Login
}

// removeDuplicates modifies a []string slice in-place removing any duplicated
// entries. This is useful to ensure that implicitly protected brnaches aren't
// duplicated. This function is largely taken from:
// https://groups.google.com/d/msg/golang-nuts/-pqkICuokio/ZfSRfU_CdmkJ
func removeDuplicates(a *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *a {
		if !found[x] {
			found[x] = true
			(*a)[j] = (*a)[i]
			j++
		}
	}
	*a = (*a)[:j]
}

// isBranchProtected checks whether the specified branch is a member of
// the specified protected branches passed in via flags.
func isBranchProtected(branch string) bool {
	for _, b := range branches {
		if branch == b {
			return true
		}
	}
	return false
}

// Get all closed issues (all PRs are issues).
func getIssues(client *github.Client, username string, page, perPage int) error {
	if lastChecked.IsZero() {
		lastChecked = time.Now()
	}

	opt := &github.IssueListOptions{
		Filter:    "created",
		State:     "closed",
		Sort:      "updated",
		Direction: "asc",
		Since:     lastChecked,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	issues, resp, err := client.Issues.List(true, opt)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		err = handleIssue(client, issue, username)
		if err != nil {
			logrus.Errorf("%v", err)
		}
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		// we probably shouldn't be polling more frequently than every
		// 5 secs, so we can rewind the clock to catch any updates that
		// may have occured while made the issues API call
		lastChecked = time.Now().Add(-5 * time.Second)
		return nil
	}

	page = resp.NextPage
	return getIssues(client, username, page, perPage)
}

// Inspect the issue and get the associated PR. Delete the associated branch is
// the PR is closed and merged.
func handleIssue(client *github.Client, issue *github.Issue, username string) error {
	if (*issue).PullRequestLinks != nil {
		pr, _, err := client.PullRequests.Get(*issue.Repository.Owner.Login, *issue.Repository.Name, *issue.Number)
		if err != nil {
			return err
		}

		if *pr.State == "closed" && *pr.Merged {
			// If the PR was made from a repository owned by the current user,
			// let's delete it.
			branch := *pr.Head.Ref
			if pr.Head.Repo == nil {
				return nil
			}
			if pr.Head.Repo.Owner == nil {
				return nil
			}
			repoOwner := *pr.Head.Repo.Owner.Login
			if pr.User.Login == nil {
				return nil
			}
			branchOwner := *pr.User.Login // the branch is owned by the user that opened the PR

			// Never delete protected branches or a branch we do not own.
			if branchOwner == username && !isBranchProtected(branch) {
				_, err := client.Git.DeleteRef(repoOwner, *pr.Head.Repo.Name, strings.Replace("heads/"+*pr.Head.Ref, "#", "%23", -1))
				// 422 is the error code for when the branch does not exist.
				if err != nil {
					if strings.Contains(err.Error(), " 422 ") {
						logrus.Infof("Branch not found: %v", err)
						return nil
					}
					return err
				}
				logrus.Infof("Branch %s on %s/%s#%v has been deleted.", branch, repoOwner, *pr.Head.Repo.Name, *pr.Number)
			}
		}
	}
	return nil
}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
