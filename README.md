# giturl
Janky tool for printing the web url for a file in a checked out git repo.

Current supports github and gitlab url formats.

# install
```
go install github.com/cnelson/giturl@latest
```

# usage
```
% giturl --help
Usage of giturl:
giturl [optional flags] <file to open>
  -branch="": The branch to use when viewing the file.  Defaults to the current working branch.
  -github-domain=: Treat this domain as a github instance.  Useful for private github installs. Can be specified more than once
  -gitlab-domain=: Treat this domain as a gitlab instance.  Useful for private github installs. Can be specified more than once

% giturl main.go:33
https://github.com/cnelson/giturl/blob/main/main.go#L33
```
