package main

import "testing"

func TestRepositoryPath(t *testing.T) {
	tests := map[string]string{
		"git@github.com:Owner/Repo.git":     "owner/repo",
		"https://github.com/Owner/Repo.git": "owner/repo",
		"ssh://git@example.com/Owner/Repo":  "owner/repo",
	}
	for input, want := range tests {
		got, err := repositoryPath(input)
		if err != nil || got != want {
			t.Errorf("repositoryPath(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}

func TestFrontendURL(t *testing.T) {
	tests := map[string]string{
		"Entire": "https://entire.io/gh/owner/repo",
		"GitLab": "https://gitlab.com/owner/repo",
	}
	for frontend, want := range tests {
		if got := frontendURL(frontend, "owner/repo"); got != want {
			t.Errorf("frontendURL(%q, %q) = %q; want %q", frontend, "owner/repo", got, want)
		}
	}
}
