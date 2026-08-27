package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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
