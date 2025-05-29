package cmd

import (
	"fmt"
	"strings"

	"github.com/carabiner-dev/vcslocator"
)

// parseLocatorToOptions parser a VCS locator into a branch options struct
func parseLocatorToOptions(lString string) (*commitOpts, error) {
	// Support specifying just repo/org
	if pts := strings.Split(lString, "/"); len(pts) == 2 {
		lString = "git+https://github.com/" + lString
	}

	// Normalize in case no schema was added
	if strings.HasPrefix(lString, "github.com/") {
		lString = "git+http://" + lString
	}

	// If not treat it as a full locator, just parse the repo slug from the path
	l := vcslocator.Locator(lString)
	components, err := l.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing %q as VCS locator: %w", lString, err)
	}
	pts := strings.Split(strings.TrimPrefix(components.RepoPath, "/"), "/")
	var owner, repo string
	repo = components.RepoPath

	if components.Hostname == "github.com" {
		if len(pts) != 2 {
			return nil, fmt.Errorf("malformed github repository repo, expecting github.com/org/repo")
		}
		owner = pts[0]
		repo = pts[1]
	}

	// We asume that any reference is the branch, tooling will fail later if it's not
	branchname := components.Branch
	if components.Commit == "" && components.Branch == "" {
		branchname = components.RefString
	}

	return &commitOpts{
		branchOpts: branchOpts{
			branch: branchname,
			repoOpts: repoOpts{
				owner: owner,
				repo:  repo,
			},
		},
		commit: components.Commit,
	}, nil
}
