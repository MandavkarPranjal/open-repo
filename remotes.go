package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type remote struct{ Name, URL string }

func gitRemotes() ([]remote, error) {
	out, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		return nil, fmt.Errorf("read git remotes: %w", err)
	}
	var result []remote
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[2] == "(fetch)" {
			result = append(result, remote{fields[0], fields[1]})
		}
	}
	return result, nil
}

func chooseRemote(remotes []remote) (remote, bool, error) {
	if len(remotes) == 1 {
		return remotes[0], true, nil
	}
	choices := make([]string, len(remotes))
	for i, r := range remotes {
		choices[i] = r.Name + " -> " + r.URL
	}
	choice, ok, err := selectItem("Select remote: ", choices, inTmuxPopup())
	if err != nil || !ok {
		return remote{}, ok, err
	}
	return remotes[indexOf(choices, choice)], true, nil
}

func repositoryPath(raw string) (string, error) {
	url := strings.TrimSpace(raw)
	var path string
	if strings.HasPrefix(url, "git@") {
		if colon := strings.IndexByte(url, ':'); colon >= 0 {
			path = url[colon+1:]
		}
	} else if scheme := strings.Index(url, "://"); scheme >= 0 {
		withoutScheme := url[scheme+3:]
		if slash := strings.IndexByte(withoutScheme, '/'); slash >= 0 {
			path = withoutScheme[slash+1:]
		}
	}
	path = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(path, ".git")))
	if path == "" {
		return "", fmt.Errorf("Could not determine repository path from remote: %s", raw)
	}
	return path, nil
}

func frontendURL(frontend, repoPath string) string {
	prefixes := map[string]string{
		"GitHub": "https://github.com/", "GitLab": "https://gitlab.com/",
		"Better Hub": "https://beta.better-hub.com/",
		"OpenGit":    "https://open-git.com/", "Codeberg": "https://codeberg.org/",
		"Gitea": "https://gitea.com/", "Entire": "https://entire.io/gh/",
		"Trylle": "https://trylle.com/", "GitCafe": "https://git.cafe/",
	}
	return prefixes[frontend] + strings.Join(strings.Fields(repoPath), "")
}

func inTmuxPopup() bool { return os.Getenv("TMUX") != "" && os.Getenv("TMUX_PANE") == "" }

func tmuxRepositoryDir() string {
	if os.Getenv("TMUX") == "" || inTmuxPopup() {
		return ""
	}
	out, err := exec.Command("tmux", "display-message", "-p", "-F", "#{pane_current_path}").Output()
	if err != nil {
		return ""
	}
	dir, _ := filepath.Abs(strings.TrimSpace(string(out)))
	return dir
}
