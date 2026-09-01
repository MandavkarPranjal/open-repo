package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type remote struct{ Name, URL string }

const commandTimeout = 2 * time.Second

var frontendPrefixes = map[string]string{
	"GitHub":     "https://github.com/",
	"GitLab":     "https://gitlab.com/",
	"Better Hub": "https://beta.better-hub.com/",
	"OpenGit":    "https://open-git.com/",
	"Codeberg":   "https://codeberg.org/",
	"Gitea":      "https://gitea.com/",
	"Entire":     "https://entire.io/gh/",
	"Trylle":     "https://trylle.com/",
	"GitCafe":    "https://git.cafe/",
}

func gitRemotes() ([]remote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "remote", "-v").Output()
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
	raw = strings.TrimSpace(raw)
	path, ok := scpLikePath(raw)
	if !ok {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return "", fmt.Errorf("could not determine repository path from remote: %s", raw)
		}
		path = parsed.EscapedPath()
	}

	path = strings.Trim(strings.TrimSpace(path), "/")
	path = strings.TrimSuffix(path, ".git")
	if path == "" {
		return "", fmt.Errorf("could not determine repository path from remote: %s", raw)
	}
	return path, nil
}

func scpLikePath(raw string) (string, bool) {
	colon := strings.IndexByte(raw, ':')
	if colon <= 0 || strings.Contains(raw[:colon], "/") || !strings.Contains(raw[:colon], "@") {
		return "", false
	}
	return raw[colon+1:], true
}

func frontendURL(frontend, repoPath string) (string, error) {
	prefix, ok := frontendPrefixes[frontend]
	if !ok {
		return "", fmt.Errorf("unknown frontend: %s", frontend)
	}
	repoPath = strings.Trim(repoPath, "/")
	if repoPath == "" || strings.ContainsAny(repoPath, " \t\n\r") {
		return "", fmt.Errorf("invalid repository path: %q", repoPath)
	}
	return prefix + repoPath, nil
}

func inTmuxPopup() bool { return os.Getenv("TMUX") != "" && os.Getenv("TMUX_PANE") == "" }

func tmuxRepositoryDir() string {
	if os.Getenv("TMUX") == "" || inTmuxPopup() {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "tmux", "display-message", "-p", "-F", "#{pane_current_path}").Output()
	if err != nil {
		return ""
	}
	dir, _ := filepath.Abs(strings.TrimSpace(string(out)))
	return dir
}
