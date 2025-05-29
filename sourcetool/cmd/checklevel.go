/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/attest"
	"github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/gh_control"
	"github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/policy"
	repomanager "github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/repo"

	"github.com/spf13/cobra"
)

type CheckLevelArgs struct {
	commit, owner, repo, branch, outputVsa, outputUnsignedVsa, useLocalPolicy string
}

// checklevelCmd represents the checklevel command
var (
	checkLevelArgs CheckLevelArgs

	checklevelCmd = &cobra.Command{
		Use:   "checklevel",
		Short: "Determines the SLSA Source Level of the repo",
		Long: `Determines the SLSA Source Level of the repo.

This is meant to be run within the corresponding GitHub Actions workflow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("vals %+v ", checkLevelArgs)
			return doCheckLevel(checkLevelArgs.commit, checkLevelArgs.owner, checkLevelArgs.repo, checkLevelArgs.branch, checkLevelArgs.outputVsa, checkLevelArgs.outputUnsignedVsa)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}

			argOpts, err := parseLocatorToOptions(args[0])
			if err != nil {
				return err
			}

			// Mix the options
			if argOpts.owner != "" {
				if checkLevelArgs.owner != "" && checkLevelArgs.owner != argOpts.owner {
					return fmt.Errorf("duplicate owner specified")
				}
				checkLevelArgs.owner = argOpts.owner
			}
			if argOpts.repo != "" {
				if checkLevelArgs.repo != "" && checkLevelArgs.repo != argOpts.repo {
					return fmt.Errorf("duplicate repository name specified")
				}
				checkLevelArgs.repo = argOpts.repo
			}
			if argOpts.commit != "" {
				if checkLevelArgs.commit != "" && checkLevelArgs.commit != argOpts.commit {
					return fmt.Errorf("duplicate commit specified")
				}
				checkLevelArgs.commit = argOpts.commit
			}
			if argOpts.branch != "" {
				if checkLevelArgs.branch != "" && checkLevelArgs.branch != argOpts.branch {
					return fmt.Errorf("duplicate branch specified")
				}
				checkLevelArgs.branch = argOpts.branch
			}
			return nil
		},
	}
)

func doCheckLevel(commit, owner, repo, branch, outputVsa, outputUnsignedVsa string) error {
	if owner == "" || repo == "" {
		log.Fatal("Must set owner and repo flags.")
	}
	var err error
	manager := repomanager.NewManager()

	if branch == "" {
		branch, err = manager.GetDefaultBranch(
			repomanager.WithRepo(repo),
			repomanager.WithOwner(owner),
		)
		if err != nil {
			return err
		}
		logrus.Infof("Read default branch %q from github", branch)
	}

	if commit == "" {
		commit, err = manager.GetLatestCommit(
			repomanager.WithOwner(owner),
			repomanager.WithRepo(repo),
			repomanager.WithBranch(branch),
		)
		if err != nil {
			return err
		}
		logrus.Infof("Read latest commit %q from github", commit)
	}

	gh_connection := gh_control.NewGhConnection(owner, repo, branch).WithAuthToken(githubToken)
	ctx := context.Background()

	controlStatus, err := gh_connection.GetControls(ctx, commit)
	if err != nil {
		log.Fatal(err)
	}

	logrus.Info("Control status:")
	spew.Dump(controlStatus)

	pol := policy.NewPolicy()
	pol.UseLocalPolicy = checkLevelProvArgs.useLocalPolicy

	verifiedLevels, policyPath, err := pol.EvaluateControl(ctx, gh_connection, controlStatus)
	if err != nil {
		fmt.Println("ERRRR")
		log.Fatal(err)
	}

	if outputUnsignedVsa != "" && outputUnsignedVsa != "-" && outputVsa != "" && outputVsa != "-" {
		fmt.Printf("LEVELS:\n")
		fmt.Print(verifiedLevels)
	}

	unsignedVsa, err := attest.CreateUnsignedSourceVsa(gh_connection, commit, verifiedLevels, policyPath)
	if err != nil {
		return fmt.Errorf("generating VSA: %w", err)
	}

	if outputUnsignedVsa != "" {
		var out io.Writer = os.Stdout
		if outputUnsignedVsa != "-" {
			out, err = os.Create(outputUnsignedVsa)
			if err != nil {
				return fmt.Errorf("opening VSA file: %w", err)
			}
			defer out.(*os.File).Close() //nolint:errcheck
		}
		if _, err := out.Write([]byte(unsignedVsa)); err != nil {
			return fmt.Errorf("writing VSA data: %w", err)
		}
		return nil
	}

	if outputVsa != "" {
		// This will output in the sigstore bundle format.
		signedVsa, err := attest.Sign(unsignedVsa)
		if err != nil {

		}
		err = os.WriteFile(outputVsa, []byte(signedVsa), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(checklevelCmd)

	// Here you will define your flags and configuration settings.

	checklevelCmd.Flags().StringVar(&checkLevelArgs.commit, "commit", "", "The commit to check.")
	checklevelCmd.Flags().StringVar(&checkLevelArgs.owner, "owner", "", "The GitHub repository owner - required.")
	checklevelCmd.Flags().StringVar(&checkLevelArgs.repo, "repo", "", "The GitHub repository name - required.")
	checklevelCmd.Flags().StringVar(&checkLevelArgs.branch, "branch", "", "The branch within the repository - required.")
	checklevelCmd.Flags().StringVar(&checkLevelArgs.outputVsa, "output_vsa", "", "The path to write a signed VSA with the determined level.")
	checklevelCmd.Flags().StringVar(&checkLevelArgs.outputUnsignedVsa, "output_unsigned_vsa", "", "The path to write an unsigned vsa with the determined level.")
	checklevelCmd.Flags().StringVar(&checkLevelArgs.useLocalPolicy, "use_local_policy", "", "UNSAFE: Use the policy at this local path instead of the official one.")

}
