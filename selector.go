package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var frontends = []string{"GitHub", "GitLab", "Better Hub", "OpenGit", "Codeberg", "Gitea", "Entire", "Trylle", "GitCafe"}

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
