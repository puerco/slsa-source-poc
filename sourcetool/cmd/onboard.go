package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/repo"
	"github.com/spf13/cobra"
)

// commitOpts
type commitOpts struct {
	branchOpts
	commit string // Commit hash (sha1)
}

func (co *commitOpts) AddFlags(cmd *cobra.Command) {
	co.branchOpts.AddFlags(cmd)
	cmd.PersistentFlags().StringVarP(
		&co.commit, "commit", "", "", "Commit hash (sha1)",
	)
}

func (co *commitOpts) Validate() error {
	errs := []error{}
	if err := co.branchOpts.Validate(); err != nil {
		errs = append(errs, err)
	}

	if len(co.commit) != 40 {
		errs = append(errs, errors.New("invalid commit digest"))
	}

	return errors.Join(errs...)
}

// Branch options (derived from repository options)
type branchOpts struct {
	branch string // Branch name
	repoOpts
}

func (bo *branchOpts) AddFlags(cmd *cobra.Command) {
	bo.repoOpts.AddFlags(cmd)
	cmd.PersistentFlags().StringVarP(
		&bo.branch, "branch", "", "", "Branch to protect, defaults to default branch (main, master, etc)",
	)
}

func (bo *branchOpts) Validate() error {
	errs := []error{}
	if err := bo.repoOpts.Validate(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// Repository options
type repoOpts struct {
	owner string // Github organization or user that owns the repo
	repo  string // Name of the repository
}

func (ro *repoOpts) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(
		&ro.owner, "owner", "", "", "The GitHub repository owner - required.",
	)
	cmd.PersistentFlags().StringVarP(
		&ro.repo, "repo", "", "", "The GitHub repository name - required.",
	)
}

// Validate checks te repossitory options
func (ro *repoOpts) Validate() error {
	errs := []error{}
	if ro.owner == "" {
		errs = append(errs, errors.New("repository owner is required"))
	}
	if ro.repo == "" {
		errs = append(errs, errors.New("repository name is required"))
	}
	return errors.Join(errs...)
}

// onboard options
type onboardOpts struct {
	branchOpts
	userForkOrg string
	enforce     bool
}

func (oo *onboardOpts) AddFlags(cmd *cobra.Command) {
	oo.branchOpts.AddFlags(cmd)

	cmd.PersistentFlags().BoolVar(
		&oo.enforce, "enforce", false, "Create enforcement rules",
	)
	cmd.PersistentFlags().StringVar(
		&oo.userForkOrg, "user_fork_org", "", "GitHub organization to look for forks of repos",
	)
}

// Validate checks the options in context with arguments
func (oo *onboardOpts) Validate() error {
	errs := []error{
		oo.branchOpts.Validate(),
	}
	return errors.Join(errs...)
}

func addOnboard(parentCmd *cobra.Command) {
	opts := &onboardOpts{}
	onboardCmd := &cobra.Command{
		Short: "onboard a new repository to SLSA Source",
		Long: `The onboard subcommand can be used to set up SLSA Source on a repository.


`,
		Use: "onboard --owner=org --repo=repository",
		// Example:           fmt.Sprintf(`%s snap --var REPO=example spec.yaml`, appname),
		SilenceUsage:  false,
		SilenceErrors: true,
		// PersistentPreRunE: initLogging,
		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				owner, repo, ok := strings.Cut(args[0], "/")
				if ok {
					if opts.repo != "" && opts.repo != repo {
						return fmt.Errorf("repository specified twice")
					}
					if opts.owner != "" && opts.owner != owner {
						return fmt.Errorf("owner specified twice")
					}
					opts.repo = repo
					opts.owner = owner
				} else {
					return fmt.Errorf("repository in argument must be an owner/repo slug")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// Validate the options
			if err := opts.Validate(); err != nil {
				return err
			}

			// At this point options are valid, no help needed.
			cmd.SilenceUsage = true

			// Create the repo manager
			manager := repo.NewManager()
			err = manager.OnboardRepository(
				repo.WithBranch(opts.branch),
				repo.WithRepo(opts.repo),
				repo.WithOwner(opts.owner),
				repo.WithEnforce(opts.enforce),
				repo.WithUserForkOrg(opts.userForkOrg),
			)
			if err != nil {
				return fmt.Errorf("onboarding repo: %w", err)
			}

			return nil
		},
	}
	opts.AddFlags(onboardCmd)
	parentCmd.AddCommand(onboardCmd)
}
