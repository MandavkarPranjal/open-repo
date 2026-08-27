package main

import (
	"errors"
	"fmt"
	"os"
)

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

	selected, ok, err := chooseRemote(remotes)
	if err != nil || !ok {
		return err
	}
	repoPath, err := repositoryPath(selected.URL)
	if err != nil {
		return err
	}

	destination, ok, err := selectItem("Open in: ", frontends, inTmuxPopup())
	if err != nil || !ok {
		return err
	}

	url := frontendURL(destination, repoPath)
	fmt.Printf("Opening: %s\n", url)
	return launchBrowser(url)
}
