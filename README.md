# ghtrim
[![Travis
CI](https://travis-ci.org/asobrien/ghtrim.svg?branch=main)](https://travis-ci.org/asobrien/ghtrim)

A GitHub Bot to automatically delete your fork's branches after a pull request
has been merged. This is fork of [ghb0t](https://github.com/jessfraz/ghb0t), that I modified
to fit my needs.

> **NOTE:** This will **never** delete a branch named "master" or "develop"
> AND will **never** delete a branch that associated with a PR *you* did not
> make. If the pull request is closed _without_ merging, it will **not** delete it.


### Why the name change?
I have too many github bots in my life, so I named this something a little
more descriptive for my needs. Alas, it's a Github branch trimmer.


### But why?
Cleanup stale branches is constantly on my todo list. So why not automate?
Thanks to [jessfraz](https://github.com/jessfraz) for all the heavy-lifting.
When I saw [this blog post](https://blog.jessfraz.com/post/personal-infrastructure/)
I thought, I can finally cross-off "clean stale branches" off my todo list.


## Usage

```
$ ghtrim -h
ghtrim - v1.0.0
  -branches string
    	protected branches, comma seperated) (default "main, master, develop")
  -d	run in debug mode
  -interval string
    	check interval (ex. 5ms, 10s, 1m, 3h) (default "30s")
  -token string
    	GitHub API token
  -v	print version and exit (shorthand)
  -version
    	print version and exit
```
