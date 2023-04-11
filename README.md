# giturl
Janky tool for printing the web url for a file in a checked out git repo.

Currently supports github and gitlab url formats.

Give a path to file a in a git repo this tool will
* Walk up the directory tree looking for the first directory containing a `.git/config`
* Parse the git config and extract the remote origin
* Determine if it is a github or gitlab url based on the url
* Generate and print the url to this file in the remote origin with optional line number.

# install
```
go install github.com/cnelson/giturl@latest
```

# usage
```
% giturl --help
Usage of giturl:
giturl [optional flags] <file to open>[:line number]
  -branch="": The branch to use when viewing the file.  Defaults to the current working branch.
  -github-domain=: Treat this domain as a github instance.  Useful for private github installs. Can be specified more than once.
  -gitlab-domain=: Treat this domain as a gitlab instance.  Useful for private gitlab installs. Can be specified more than once.

% giturl main.go:33
https://github.com/cnelson/giturl/blob/main/main.go#L33
```
