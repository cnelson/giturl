package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/namsral/flag"
	"gopkg.in/ini.v1"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func httpUrl(url string) (string, error) {
	if strings.HasPrefix(url, "git@") {
		url = strings.Replace(url, ":", "/", 1)
		url = strings.Replace(url, "git@", "https://", 1)
	}

	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return url, nil
	} else {
		return url, fmt.Errorf("Can't convert %s to https url", url)
	}

}

func getCurrentBranch(repoDir string) (branch string, err error) {
	cmd := exec.Command(
		"git",
		"branch",
		"--show-current",
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = repoDir
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Unable to determine branch name: %s\n%s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func findGitUrl(fn string, branch string, providers map[string]string) (result string, err error) {
	// walk backwards up the file tree looking for a .git/config
	dir, _ := filepath.Split(fn)
	var fh *os.File
	for len(dir) > 0 && dir != "/" {
		fh, err = os.Open(filepath.Join(dir, ".git/config"))
		if err == nil {
			defer fh.Close()
			break
		}
		dir, _ = filepath.Split(strings.TrimSuffix(dir, "/"))
	}

	// didn't find a git config
	if fh == nil {
		return "", errors.New("Unable to locate git config")
	}

	// have one, try to parse it
	cfg, err := ini.Load(fh)
	if err != nil {
		return "", err
	}

	// look for the first remote section
	var remoteSection *ini.Section
	for _, s := range cfg.Sections() {
		if strings.HasPrefix(s.Name(), "remote") {
			remoteSection = s
			break
		}
	}
	if remoteSection == nil || !remoteSection.HasKey("url") {
		return "", fmt.Errorf("Couldn't find remote section or url in %s", fh.Name())
	}

	baseUrl, err := httpUrl(remoteSection.Key("url").String())
	if err != nil {
		return "", err
	}
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	baseUrl = strings.TrimSuffix(baseUrl, ".git")

	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}
	provider, ok := providers[parsedUrl.Hostname()]
	if provider == "" {
		return "", fmt.Errorf("Unable to determine provider for url: %s", parsedUrl.Hostname())
	}

	prefixes := map[string]string{
		"gitlab": "/-/blob",
		"github": "/blob",
	}

	// assume gitlab unless otherwise
	prefix, ok := prefixes[provider]
	if !ok {
		return "", fmt.Errorf("Unknown provider %s", provider)
	}

	// if they didn't give us a branch, then figure it out
	if branch == "" {
		branch, err = getCurrentBranch(dir)
		if err != nil {
			return "", err
		}
	}

	result = baseUrl + fmt.Sprintf("%s/%s/", prefix, branch) + strings.TrimPrefix(fn, dir)
	return
}

type domainList []string

func (d *domainList) String() string {
	if len(*d) == 0 {
		return ""
	} else {
		return strings.Join(*d, ", ")
	}
}

func (d *domainList) Set(value string) error {
	*d = append(*d, value)
	return nil
}
func main() {
	var branch string

	var githubDomains domainList
	var gitlabDomains domainList

	_, cmd := filepath.Split(os.Args[0])

	fs := flag.NewFlagSetWithEnvPrefix(cmd, "GITOPENER", 0)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", cmd)
		fmt.Fprintf(os.Stderr, "%s [optional flags] <file to open>\n", cmd)
		fs.PrintDefaults()
	}

	fs.StringVar(&branch, "branch", "", "The branch to use when viewing the file.  Defaults to the current working branch.")
	fs.Var(&githubDomains, "github-domain", `Treat this domain as a github instance.  Useful for private github installs. Can be specified more than once`)
	fs.Var(&gitlabDomains, "gitlab-domain", `Treat this domain as a gitlab instance.  Useful for private github installs. Can be specified more than once`)

	if fs.Parse(os.Args[1:]) != nil {
		os.Exit(1)
	} else {
		// we need at least one other arg
		if fs.NArg() != 1 {
			fs.Usage()
			os.Exit(1)
		}
	}

	// load any custom domains they gave us
	providerMapping := map[string]string{
		"github.com": "github",
		"gitlab.com": "gitlab",
	}
	for _, v := range githubDomains {
		providerMapping[v] = "github"
	}
	for _, v := range gitlabDomains {
		providerMapping[v] = "gitlab"
	}

	// check for line number at the end
	targetFilename, _ := filepath.Abs(fs.Arg(0))
	targetLinenum := 0

	if idx := strings.LastIndex(targetFilename, ":"); idx > 0 {
		if z, err := strconv.Atoi(targetFilename[idx+1:]); err == nil {
			targetLinenum = z
			targetFilename = targetFilename[:idx]
		} else {
			fmt.Fprintf(os.Stderr, "Invalid line number: %+v, %s\n", os.Args, err)
			fs.Usage()
			os.Exit(2)
		}
	}

	// find the url
	weburl, err := findGitUrl(targetFilename, branch, providerMapping)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		fs.Usage()
		os.Exit(2)
	}

	if targetLinenum > 0 {
		weburl += fmt.Sprintf("#L%d", targetLinenum)
	}

	fmt.Println(weburl)

}
