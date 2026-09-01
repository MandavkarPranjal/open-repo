package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

func launchBrowser(url string) error {
	name, args, ok, err := configuredBrowser()
	if err != nil {
		return err
	}
	if ok {
		return startDetached(name, append(args, url)...)
	}

	switch runtime.GOOS {
	case "linux":
		for _, opener := range []struct {
			name string
			args []string
		}{
			{"gio", []string{"open", url}}, {"xdg-open", []string{url}},
		} {
			if _, err := exec.LookPath(opener.name); err == nil {
				return startDetached(opener.name, opener.args...)
			}
		}
	case "darwin":
		return startDetached("open", url)
	case "windows":
		return startDetached("rundll32", "url.dll,FileProtocolHandler", url)
	}
	return fmt.Errorf("no browser opener available for %s", runtime.GOOS)
}

func configuredBrowser() (string, []string, bool, error) {
	if value := os.Getenv("BROWSER"); value != "" {
		return splitBrowserCommand(value)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", nil, false, nil
	}
	file, err := os.Open(filepath.Join(configDir, "open-repo", "config.toml"))
	if os.IsNotExist(err) {
		return "", nil, false, nil
	}
	if err != nil {
		return "", nil, false, fmt.Errorf("read browser configuration: %w", err)
	}
	defer func() { _ = file.Close() }()

	var config struct {
		Browser string `toml:"browser"`
	}
	if err := toml.NewDecoder(file).Decode(&config); err != nil {
		return "", nil, false, fmt.Errorf("parse browser configuration: %w", err)
	}
	if config.Browser == "" {
		return "", nil, false, nil
	}
	return splitBrowserCommand(config.Browser)
}

func splitBrowserCommand(command string) (string, []string, bool, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil, false, fmt.Errorf("browser command is empty")
	}
	return parts[0], parts[1:], true, nil
}

func startDetached(name string, args ...string) error {
	if runtime.GOOS != "windows" {
		if setsid, err := exec.LookPath("setsid"); err == nil {
			args = append([]string{name}, args...)
			name = setsid
		}
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	return cmd.Start()
}
