package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/gh_control"
	"github.com/spf13/cobra"
)

type controlsOpts struct {
	repoOpts
}

// Validate checks the options in context with arguments
func (co *controlsOpts) Validate() error {
	errs := []error{
		co.repoOpts.Validate(),
	}
	return errors.Join(errs...)
}

// AddFlags adds the subcommands flags
func (co *controlsOpts) AddFlags(cmd *cobra.Command) {
	co.repoOpts.AddFlags(cmd)
}

func addControls(parentCmd *cobra.Command) {
	controlsCmd := &cobra.Command{
		Short: "Commands to interact with the repository controls",
		Use:   "controls",
		Long: `The controls subcommand can be used check and set up the controls in the repository.

`,
		SilenceUsage:  false,
		SilenceErrors: true,
		// PersistentPreRunE: initLogging,
	}
	// opts.AddFlags(onboardCmd)
	addControlsStatus(controlsCmd)
	parentCmd.AddCommand(controlsCmd)
}

type controlsStatusOpts struct {
	branchOpts
	commit string
}

// Validate checks the options in context with arguments
func (co *controlsStatusOpts) Validate() error {
	errs := []error{
		co.repoOpts.Validate(),
	}

	if co.commit != "" && co.commit != "HEAD" && len(co.commit) != 40 {
		errs = append(errs, fmt.Errorf("invalid commit"))
	}
	return errors.Join(errs...)
}

// AddFlags adds the subcommands flags
func (co *controlsStatusOpts) AddFlags(cmd *cobra.Command) {
	co.repoOpts.AddFlags(cmd)

	cmd.PersistentFlags().StringVarP(
		&co.commit, "commit", "", "", "Commit to check",
	)
}

func addControlsStatus(parentCmd *cobra.Command) {
	opts := &controlsStatusOpts{}
	onboardCmd := &cobra.Command{
		Short: "Reports the status of the repository controls",
		Use:   "status [flags] org/repo",
		// Example:           fmt.Sprintf(`%s snap --var REPO=example spec.yaml`, appname),
		SilenceUsage:  false,
		SilenceErrors: true,
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

			githubToken = os.Getenv("GITHUB_TOKEN")
			gh_connection := gh_control.NewGhConnection(opts.owner, opts.repo, opts.branch).WithAuthToken(githubToken)
			ctx := context.Background()

			controlStatus, err := gh_connection.GetControls(ctx, opts.commit)
			if err != nil {
				return fmt.Errorf("reading repo controls: %w", err)
			}

			fmt.Printf("%+v", controlStatus)

			return nil
		},
	}
	opts.AddFlags(onboardCmd)
	parentCmd.AddCommand(onboardCmd)
}

func init() {
	addControls(rootCmd)
}
