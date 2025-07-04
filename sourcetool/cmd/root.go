/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/attest"
)

var (
	githubToken string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "sourcetool",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
)

func getVerifier(vo *verifierOptions) attest.Verifier {
	options := attest.DefaultVerifierOptions
	if vo.expectedIssuer != "" {
		options.ExpectedIssuer = vo.expectedIssuer
	}
	if vo.expectedSan != "" {
		options.ExpectedSan = vo.expectedSan
	}
	return attest.NewBndVerifier(options)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&githubToken, "github_token", "", "the github token to use for auth")

	addCheckLevel(rootCmd)
	addVerifyCommit(rootCmd)
	addStatus(rootCmd)
	addSetup(rootCmd)
	addAudit(rootCmd)
	addProv(rootCmd)
}
