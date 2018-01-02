# ghedgetrim

This is  fork of [ghb0t](https://github.com/jessfraz/ghb0t), that I modified
to fit my needs.

### Why the name change?
I have too many github bots in my life, so I named this something a little
more descriptive for my needs. Alas, it's a Github hedge trimmer.

### But why?
Cleanup stale branches is constantly on my todo list. So why not automate?
Thanks to [jessfraz](https://github.com/jessfraz) for all the heavy-lifting.
When I saw [this blog post](https://blog.jessfraz.com/post/personal-infrastructure/)
I thought, I can finally cross-off "clean stale branches" off my todo list.

---

[![Travis CI](https://travis-ci.org/asobrien/ghedgetrim.svg?branch=master)](https://travis-ci.org/asobrien/ghedgetrim)

A GitHub Bot to automatically delete your fork's branches after a pull request
has been merged.

> **NOTE:** This will **never** delete a branch named "master" AND will
**never** delete a branch that is not owned by the current authenticated user.
If the pull request is closed _without_ merging, it will **not** delete it.

## Usage

```
$ ghb0t -h
ghb0t - v0.1.0
  -d    run in debug mode
  -seconds int
        seconds to wait before checking for new events (default 30)
  -token string
        GitHub API token
  -v    print version and exit (shorthand)
  -version
        print version and exit
```
