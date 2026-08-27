package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type remote struct {
	Name string
	URL  string
}

var frontends = []string{"GitHub", "Better Hub", "OpenGit", "Codeberg", "Gitea", "Entire", "Trylle", "GitCafe"}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if dir := tmuxRepositoryDir(); dir != "" {
		if err := os.Chdir(dir); err != nil {
			return fmt.Errorf("change to tmux repository directory: %w", err)
		}
	}

	remotes, err := gitRemotes()
	if err != nil {
		return err
	}
	if len(remotes) == 0 {
		return errors.New("No git remotes found")
	}

	var selected remote
	if len(remotes) == 1 {
		selected = remotes[0]
	} else {
		choices := make([]string, len(remotes))
		for i, r := range remotes {
			choices[i] = r.Name + " -> " + r.URL
		}
		choice, ok, err := selectItem("Select remote: ", choices, inTmuxPopup())
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		selected = remotes[indexOf(choices, choice)]
	}

	repoPath, err := repositoryPath(selected.URL)
	if err != nil {
		return err
	}
	destination, ok, err := selectItem("Open in: ", frontends, inTmuxPopup())
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	url := frontendURL(destination, repoPath)
	fmt.Printf("Opening: %s\n", url)
	return launchBrowser(url)
}

func gitRemotes() ([]remote, error) {
	out, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		return nil, fmt.Errorf("read git remotes: %w", err)
	}
	var result []remote
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[2] == "(fetch)" {
			result = append(result, remote{Name: fields[0], URL: fields[1]})
		}
	}
	return result, nil
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
	path = strings.TrimSuffix(path, ".git")
	path = strings.ToLower(strings.TrimSpace(path))
	if path == "" {
		return "", fmt.Errorf("Could not determine repository path from remote: %s", raw)
	}
	return path, nil
}

func frontendURL(frontend, repoPath string) string {
	prefixes := map[string]string{
		"GitHub": "https://github.com/", "Better Hub": "https://beta.better-hub.com/",
		"OpenGit": "https://open-git.com/", "Codeberg": "https://codeberg.org/",
		"Gitea": "https://gitea.com/", "Entire": "https://entire.io/gh/",
		"Trylle": "https://trylle.com/", "GitCafe": "https://git.cafe/",
	}
	return prefixes[frontend] + strings.Join(strings.Fields(repoPath), "")
}

func selectItem(prompt string, choices []string, popup bool) (string, bool, error) {
	if fzf, err := exec.LookPath("fzf"); err == nil {
		args := []string{"--prompt=" + prompt, "--reverse"}
		if !popup {
			args = append(args, "--height=40%")
		}
		cmd := exec.Command(fzf, args...)
		cmd.Stdin = strings.NewReader(strings.Join(choices, "\n") + "\n")
		out, err := cmd.Output()
		if err != nil {
			if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
				return "", false, nil
			}
			return "", false, fmt.Errorf("run fzf: %w", err)
		}
		return strings.TrimSpace(string(out)), true, nil
	}

	// A portable fallback for machines without fzf.
	fmt.Println(prompt)
	for i, choice := range choices {
		fmt.Printf("  %d) %s\n", i+1, choice)
	}
	fmt.Print("Choose a number (or press Enter to cancel): ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", false, nil
	}
	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(choices) {
		return "", false, errors.New("invalid selection")
	}
	return choices[n-1], true, nil
}

func indexOf(items []string, value string) int {
	for i, item := range items {
		if item == value {
			return i
		}
	}
	return 0
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

func launchBrowser(url string) error {
	if runtime.GOOS == "linux" {
		if desktop, _ := defaultLinuxBrowser(); desktop == "app.zen_browser.zen.desktop" {
			if _, err := exec.LookPath("uwsm-app"); err == nil {
				return startDetached("systemd-run", "--user", "--quiet", "--collect", "--unit=open-repo-browser", "uwsm-app", "--", "flatpak", "run", "app.zen_browser.zen", url)
			}
		}
		for _, opener := range []struct {
			name string
			args []string
		}{
			{"omarchy-launch-browser", []string{url}}, {"gio", []string{"open", url}}, {"xdg-open", []string{url}},
		} {
			if _, err := exec.LookPath(opener.name); err == nil {
				return startDetached(opener.name, opener.args...)
			}
		}
	} else if runtime.GOOS == "darwin" {
		return startDetached("open", url)
	} else if runtime.GOOS == "windows" {
		return startDetached("rundll32", "url.dll,FileProtocolHandler", url)
	}
	return fmt.Errorf("no browser opener available for %s", runtime.GOOS)
}

func defaultLinuxBrowser() (string, error) {
	cmd := exec.Command("xdg-settings", "get", "default-web-browser")
	cmd.Env = make([]string, 0, len(os.Environ()))
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "BROWSER=") {
			cmd.Env = append(cmd.Env, value)
		}
	}
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func startDetached(name string, args ...string) error {
	if runtime.GOOS != "windows" {
		if setsid, err := exec.LookPath("setsid"); err == nil && name != "systemd-run" {
			args = append([]string{name}, args...)
			name = setsid
		}
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	return cmd.Start()
}
