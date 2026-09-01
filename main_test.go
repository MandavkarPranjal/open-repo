package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRepositoryPath(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"scp-like GitHub", "git@github.com:Owner/Repo.git", "Owner/Repo"},
		{"HTTPS with credentials", "https://token@github.com/Owner/Repo.git", "Owner/Repo"},
		{"SSH port", "ssh://git@example.com:2222/Owner/Repo.git", "Owner/Repo"},
		{"git protocol", "git://example.com/org/subgroup/repo.git", "org/subgroup/repo"},
		{"subgroup trailing slash", "https://codeberg.org/Org/Subgroup/Repo.git/", "Org/Subgroup/Repo"},
		{"case is preserved", "https://gitea.example/Owner/MixedCase.git", "Owner/MixedCase"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repositoryPath(tt.input)
			if err != nil || got != tt.want {
				t.Errorf("repositoryPath(%q) = %q, %v; want %q", tt.input, got, err, tt.want)
			}
		})
	}
}

func TestFZFArgsKeepPromptSeparate(t *testing.T) {
	prompt := "Select --query=owner/repo: "
	args := fzfArgs(prompt, false)
	if args[0] != "--prompt" || args[1] != prompt {
		t.Fatalf("fzfArgs() starts with %q; want separate prompt flag and value", args[:2])
	}
}

func TestParseOptions(t *testing.T) {
	var stdout, stderr bytes.Buffer
	opts, runProgram, err := parseOptions([]string{"--remote", "origin", "--frontend", "GitHub", "--print-url"}, &stdout, &stderr)
	if err != nil || !runProgram || opts.remote != "origin" || opts.frontend != "GitHub" || !opts.printURL {
		t.Fatalf("parseOptions() = %#v, %t, %v", opts, runProgram, err)
	}

	stdout.Reset()
	_, runProgram, err = parseOptions([]string{"--version"}, &stdout, &stderr)
	if err != nil || runProgram || stdout.String() != version+"\n" {
		t.Fatalf("version handling = %q, %t, %v", stdout.String(), runProgram, err)
	}

	stderr.Reset()
	_, runProgram, err = parseOptions([]string{"--help"}, &stdout, &stderr)
	if err != nil || runProgram || !bytes.Contains(stderr.Bytes(), []byte("--remote NAME")) {
		t.Fatalf("help handling = %q, %t, %v", stderr.String(), runProgram, err)
	}
}

func TestConfiguredBrowserReadsTOML(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("BROWSER", "")
	dir := filepath.Join(configDir, "open-repo")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := "# Browser command\nbrowser = \"firefox --private-window\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	name, args, ok, err := configuredBrowser()
	if err != nil || !ok || name != "firefox" || !reflect.DeepEqual(args, []string{"--private-window"}) {
		t.Fatalf("configuredBrowser() = %q, %q, %t, %v", name, args, ok, err)
	}
}

func TestFrontendURL(t *testing.T) {
	tests := []struct {
		frontend, want string
	}{
		{"Entire", "https://entire.io/gh/owner/repo"},
		{"GitLab", "https://gitlab.com/owner/repo"},
	}
	for _, tt := range tests {
		got, err := frontendURL(tt.frontend, "owner/repo")
		if err != nil || got != tt.want {
			t.Errorf("frontendURL(%q, %q) = %q, %v; want %q", tt.frontend, "owner/repo", got, err, tt.want)
		}
	}
	if _, err := frontendURL("GitHbu", "owner/repo"); err == nil {
		t.Error("frontendURL accepted an unknown frontend")
	}
}
