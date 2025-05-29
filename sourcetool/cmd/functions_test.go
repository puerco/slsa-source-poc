package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLocatorToOptions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		sut      string
		expected *commitOpts
		mustErr  bool
	}{
		{
			"org/repo", "sigstore/cosign", &commitOpts{
				branchOpts: branchOpts{
					repoOpts: repoOpts{owner: "sigstore", repo: "cosign"},
				},
			}, false,
		},
		{
			"github.com/org/repo", "github.com/sigstore/cosign", &commitOpts{
				branchOpts: branchOpts{
					repoOpts: repoOpts{owner: "sigstore", repo: "cosign"},
				},
			}, false,
		},
		{
			"vcs-no-ref", "git+https://github.com/sigstore/cosign", &commitOpts{
				branchOpts: branchOpts{
					repoOpts: repoOpts{owner: "sigstore", repo: "cosign"},
				},
			}, false,
		},
		{
			"vcs-commit", "git+https://github.com/sigstore/cosign@e3666897979f2992d2fe8ff24c066711db08c0f6", &commitOpts{
				branchOpts: branchOpts{
					repoOpts: repoOpts{owner: "sigstore", repo: "cosign"},
				},
				commit: "e3666897979f2992d2fe8ff24c066711db08c0f6",
			}, false,
		},
		{
			"vcs-nbranch", "git+https://github.com/sigstore/cosign@feature-1.8", &commitOpts{
				branchOpts: branchOpts{
					repoOpts: repoOpts{owner: "sigstore", repo: "cosign"}, branch: "feature-1.8",
				},
			}, false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts, err := parseLocatorToOptions(tc.sut)
			if tc.mustErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected.owner, opts.owner)
			require.Equal(t, tc.expected.repo, opts.repo)
			require.Equal(t, tc.expected.branch, opts.branch)
			require.Equal(t, tc.expected.commit, opts.commit)
		})
	}
}
