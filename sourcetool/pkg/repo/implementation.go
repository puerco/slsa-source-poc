package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/carabiner-dev/github"
	gogit "github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	ghcontrol "github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/gh_control"
	"github.com/slsa-framework/slsa-source-poc/sourcetool/pkg/policy"
	kgithub "sigs.k8s.io/release-sdk/github"
)

const (
	tokenVar = "GITHUB_TOKEN"

	// FIXME: These are the real ones, uncomment before commit
	// SlsaSourceRepo = "slsa-source-poc"
	// SlsaSourceOrg  = "slsa-framework"
	SlsaSourceRepo = "slsa-poc-test"
	SlsaSourceOrg  = "carabiner-dev"

	workflowPath   = ".github/workflows/compute_slsa_source.yaml"
	workflowSource = "git+https://github.com/slsa-"

	policySource = "git+https://"
)

// repoManagerImplementation
type repoManagerImplementation interface {
	EnsureDefaults(opts *Options) error
	VerifyOptions(*Options) error
	CreateRepoRuleset(*Options) error
	CreateWorkflowPR(*Options) error
	CreatePolicyPR(*Options) error
	getBranchData(*Options) error
	getLatestCommit(opts *Options) (string, error)
}

type defaultrepoManagerImplementation struct{}

// EnsureBranch makes sure the manager has a defined branch, looking up the
// default if it needs to
func (impl *defaultrepoManagerImplementation) EnsureDefaults(opts *Options) error {
	if t := os.Getenv(tokenVar); t == "" {
		return fmt.Errorf("$%s environment variable not set", tokenVar)
	}

	// Load the default branch unless we're using a custom branch
	if err := impl.getBranchData(opts); err != nil {
		return err
	}

	// Load the token user to use as source org
	if err := getUserData(opts); err != nil {
		return err
	}

	// Output the computed defaults
	logrus.Infof("We will create branches based on forks from %q", opts.UserForkOrg)
	logrus.Infof("Using default branch %q", opts.Branch)
	return nil
}

// getHeadCommit returns the latest commit hash of the specified branch
func (impl *defaultrepoManagerImplementation) getLatestCommit(opts *Options) (string, error) {
	if opts.Branch == "" {
		return "", fmt.Errorf("no branch specified")
	}

	// Fetch the default branch
	client, err := github.NewClient()
	if err != nil {
		return "", fmt.Errorf("creating GitHub client: %w", err)
	}

	var branchdata = struct {
		CommitData struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}{}

	res, err := client.Call(
		context.Background(), http.MethodGet,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", opts.Owner, opts.Repo, opts.Branch), nil,
	)
	if err != nil {
		return "", fmt.Errorf("fetching branch data: %w", err)
	}
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading data: %w", err)
	}

	if err := json.Unmarshal(data, &branchdata); err != nil {
		return "", fmt.Errorf("unmarshaling branch data: %w", err)
	}

	if branchdata.CommitData.SHA == "" {
		return "", fmt.Errorf("unable to latest commit SHA")
	}

	return branchdata.CommitData.SHA, nil
}

func (impl *defaultrepoManagerImplementation) getBranchData(opts *Options) error {
	if opts.Branch != "" {
		return nil
	}

	// Fetch the default branch
	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("creating GitHub client: %w", err)
	}

	var repodata = struct {
		DefaultBranch string `json:"default_branch"`
	}{}

	res, err := client.Call(
		context.Background(), http.MethodGet,
		fmt.Sprintf("https://api.github.com/repos/%s/%s", opts.Owner, opts.Repo), nil,
	)
	if err != nil {
		return fmt.Errorf("fetching repository data: %s", err)
	}
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading data: %w", err)
	}

	if err := json.Unmarshal(data, &repodata); err != nil {
		return fmt.Errorf("unmarshaling repo data: %w", err)
	}

	if repodata.DefaultBranch == "" {
		return fmt.Errorf("unable to read default branch")
	}

	opts.Branch = repodata.DefaultBranch

	return nil
}

func getUserData(opts *Options) error {
	if opts.UserForkOrg != "" {
		return nil
	}
	// Fetch the default branch
	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("creating GitHub client: %w", err)
	}

	// Call the api to get the user's data
	res, err := client.Call(
		context.Background(), http.MethodGet,
		"https://api.github.com/user", nil,
	)
	if err != nil {
		return fmt.Errorf("fetching user data: %s", err)
	}
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading user data: %w", err)
	}

	var userdata = struct {
		Login string `json:"login"`
	}{}

	if err := json.Unmarshal(data, &userdata); err != nil {
		return fmt.Errorf("unmarshaling repo data: %w", err)
	}

	if userdata.Login == "" {
		return fmt.Errorf("unable to read user login for the token")
	}

	opts.UserForkOrg = userdata.Login
	return nil
}

// VerifyOptions checks options are in good shape to run
func (impl *defaultrepoManagerImplementation) VerifyOptions(opts *Options) error {
	errs := []error{}
	if opts.Repo == "" {
		errs = append(errs, errors.New("no repository name defined"))
	}

	if opts.Owner == "" {
		errs = append(errs, errors.New("no repository owner defined"))
	}

	if t := os.Getenv(tokenVar); t == "" {
		errs = append(errs, fmt.Errorf("$%s environment variable not set", tokenVar))
	}

	if opts.Enforce {
		client, err := github.NewClient()
		if err != nil {
			errs = append(errs, fmt.Errorf("creating GitHub client: %w", err))
		} else {
			scopes, err := client.TokenScopes()
			if err == nil {
				if !slices.Contains(scopes, "admin:write") {
					errs = append(errs, fmt.Errorf(`unable to create enforcing branch rules, token needs "Administration" repository permissions (write)`))
				}
			} else {
				errs = append(errs, fmt.Errorf("checking token scopes: %w", err))
			}
		}
	}

	return errors.Join(errs...)
}

func (impl *defaultrepoManagerImplementation) CreateRepoRuleset(*Options) error {
	return nil
}

func (impl *defaultrepoManagerImplementation) CreateWorkflowPR(opts *Options) error {
	// Branchname to be created on the user's fork
	branchname := fmt.Sprintf("slsa-source-workflow-%d", time.Now().Unix())

	// Check Environment
	gh := kgithub.New()

	userForkOrg := opts.UserForkOrg
	userForkRepo := opts.Repo // For now we only support forks with the same name

	if err := kgithub.VerifyFork(
		branchname, userForkOrg, userForkRepo, opts.Owner, opts.Repo,
	); err != nil {
		return fmt.Errorf(
			"while checking fork of %s/%s in %s: %w ",
			opts.Owner, opts.Repo, opts.UserForkOrg, err,
		)
	}

	// Clone the repository being onboarded
	gitCloneOpts := &gogit.CloneOptions{Depth: 1}
	repo, err := kgithub.PrepareFork(
		branchname, opts.Owner, opts.Repo,
		userForkOrg, userForkRepo,
		opts.UseSSH, opts.UpdateRepo, gitCloneOpts,
	)
	if err != nil {
		return fmt.Errorf("while preparing the repository fork: %w", err)
	}

	defer func() {
		repo.Cleanup() //nolint:errcheck
	}()

	// Create the workflow file here
	fullPath := filepath.Join(repo.Dir(), workflowPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.FileMode(0o755)); err != nil {
		return fmt.Errorf("creating workflow directory: %w", err)
	}

	// Write the workflow file to disk
	if err := os.WriteFile(fullPath, []byte(workflowData), os.FileMode(0o644)); err != nil {
		return fmt.Errorf("writing workflow data to disk: %w", err)
	}

	// add the modified manifest to staging
	logrus.Debugf("Adding %s to staging area", workflowPath)
	if err := repo.Add(workflowPath); err != nil {
		return fmt.Errorf("adding workflow file to staging area: %w", err)
	}

	commitMessage := "Add SLSA Source attesting workflow"

	// Create the commit
	if err := repo.UserCommit(commitMessage); err != nil {
		return fmt.Errorf("committing changes to workflow: %w", err)
	}

	// Push commit to branch in the user's fork
	logrus.Infof("Pushing workflow commit to %s/%s", userForkOrg, userForkRepo)
	if err := repo.PushToRemote(kgithub.UserForkName, branchname); err != nil {
		return fmt.Errorf("pushing %s to %s/%s: %w", kgithub.UserForkName, userForkOrg, userForkRepo, err)
	}

	prBody := `This pull request adds a workflow to the repository to attest the
SLSA Source compliance on every push.
`
	// Create the Pull Request
	pr, err := gh.CreatePullRequest(
		opts.Owner, opts.Repo, opts.Branch,
		fmt.Sprintf("%s:%s", userForkOrg, branchname),
		commitMessage, prBody, false,
	)
	if err != nil {
		return fmt.Errorf("creating the pull request in %s: %w", opts.Owner, err)
	}
	logrus.Infof(
		"Successfully created PR: %s%s/%s/pull/%d",
		kgithub.GitHubURL, opts.Owner, opts.Repo, pr.GetNumber(),
	)

	// Success!
	return nil
}

// CreatePolicyPR creates a pull request to push the policy
func (impl *defaultrepoManagerImplementation) CreatePolicyPR(opts *Options) error {
	// Branchname to be created on the user's fork
	branchname := fmt.Sprintf("slsa-source-policy-%d", time.Now().Unix())

	gh := kgithub.New()

	userForkOrg := opts.UserForkOrg
	userForkRepo := SlsaSourceRepo // For now we only support forks with the same name

	// Check the user has a fork of the slsa repo
	if err := kgithub.VerifyFork(
		branchname, userForkOrg, userForkRepo, SlsaSourceOrg, SlsaSourceRepo,
	); err != nil {
		return fmt.Errorf(
			"while checking fork of %s/%s in %s: %w ",
			SlsaSourceOrg, SlsaSourceRepo, userForkOrg, err,
		)
	}

	// Clone the slsa repo
	gitCloneOpts := &gogit.CloneOptions{Depth: 1}
	repo, err := kgithub.PrepareFork(
		branchname, SlsaSourceOrg, SlsaSourceRepo,
		userForkOrg, userForkRepo,
		opts.UseSSH, opts.UpdateRepo, gitCloneOpts,
	)
	if err != nil {
		return fmt.Errorf("while preparing Slsa Source fork: %w", err)
	}

	defer func() {
		repo.Cleanup() //nolint:errcheck
	}()

	// Create the policy in the local clone
	ghc := ghcontrol.NewGhConnection(opts.Owner, opts.Repo, opts.Branch).WithAuthToken(os.Getenv(tokenVar))
	outpath, err := policy.CreateLocalPolicy(context.Background(), ghc, repo.Dir())
	if err != nil {
		return fmt.Errorf("creating local policy: %w", err)
	}

	// add the modified manifest to staging
	logrus.Debugf("Adding %s to staging area", outpath)
	if err := repo.Add(strings.TrimPrefix(strings.TrimPrefix(outpath, repo.Dir()), "/")); err != nil {
		return fmt.Errorf("adding new policy file to staging area: %w", err)
	}

	commitMessage := fmt.Sprintf("Add %s/%s SLSA Source policy file", opts.Owner, opts.Repo)

	// Commit files
	if err := repo.UserCommit(commitMessage); err != nil {
		return fmt.Errorf("creating commit in %s/%s: %w", SlsaSourceOrg, SlsaSourceRepo, err)
	}

	// Push to fork
	logrus.Infof("Pushing policy commit to %s/%s", userForkOrg, userForkRepo)
	if err := repo.PushToRemote(kgithub.UserForkName, branchname); err != nil {
		return fmt.Errorf("pushing %s to %s/%s: %w", kgithub.UserForkName, userForkOrg, userForkRepo, err)
	}

	prBody := fmt.Sprintf(`This pull request adds the SLSA source policy for github.com/%s/%s`, opts.Owner, opts.Repo)

	// Create the Pull Request
	pr, err := gh.CreatePullRequest(
		SlsaSourceOrg, SlsaSourceRepo, "main",
		fmt.Sprintf("%s:%s", userForkOrg, branchname),
		commitMessage, prBody, false,
	)
	if err != nil {
		logrus.Infof("%+v", err)
		return fmt.Errorf("creating the policy PR in %s/%s: %w", SlsaSourceOrg, SlsaSourceRepo, err)
	}
	logrus.Infof(
		"Successfully created PR: %s%s/%s/pull/%d",
		kgithub.GitHubURL, SlsaSourceOrg, SlsaSourceRepo, pr.GetNumber(),
	)

	// Success!
	return nil
}

type ruleData struct {
	Type       string         `json:"type"`
	Parameters map[string]any `json:"parameters"`
}

// getCurrentRules retrieves the existing branch rules in the repository
func (impl *defaultrepoManagerImplementation) getCurrentRules(opts *Options) ([]ruleData, error) {
	rules := []ruleData{}
	client, err := github.NewClient()
	if err != nil {
		return nil, fmt.Errorf("creating GitHub client: %w", err)
	}

	res, err := client.Call(
		context.Background(), http.MethodGet,
		fmt.Sprintf("/repos/%s/%s/rules/branches/%s", opts.Owner, opts.Repo, opts.Branch), nil,
	)
	if err != nil {
		return nil, fmt.Errorf("calling API to retrieve rules: %w", err)
	}
	defer res.Body.Close() //nolint:errcheck

	// Read the response data
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response data: %w", err)
	}

	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("unmarshaling branch protection data: %w", err)
	}
	return rules, nil
}
