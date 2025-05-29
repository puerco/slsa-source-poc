package repo

import "fmt"

func NewManager() *Manager {
	return &Manager{
		impl: &defaultrepoManagerImplementation{},
	}
}

// Manager is a tool to manage the repository interactions the SLSA source
// tooling requires.
type Manager struct {
	impl repoManagerImplementation
}

// OnboardRepository configures a repository to set up the SLSA source tools.
func (m *Manager) OnboardRepository(funcs ...ooFn) error {
	opts := &Options{}
	for _, f := range funcs {
		if err := f(opts); err != nil {
			return err
		}
	}

	if err := m.impl.EnsureDefaults(opts); err != nil {
		return fmt.Errorf("ensuring runtime defaults: %w", err)
	}

	if err := m.impl.VerifyOptions(opts); err != nil {
		return fmt.Errorf("verifying options: %w", err)
	}

	// if err := m.impl.CreateRepoRuleset(opts); err != nil {
	// 	return fmt.Errorf("creating rules in the repository: %w", err)
	// }

	// if err := m.impl.CreateWorkflowPR(opts); err != nil {
	// 	return fmt.Errorf("opening SLSA source workflow pull request: %w", err)
	// }

	if err := m.impl.CreatePolicyPR(opts); err != nil {
		return fmt.Errorf("opening the policy pull request: %w", err)
	}

	return nil
}

// GetDefaultBranch returns the default branch of a repository
func (m *Manager) GetDefaultBranch(funcs ...ooFn) (string, error) {
	opts := &Options{}
	for _, f := range funcs {
		if err := f(opts); err != nil {
			return "", err
		}
	}

	if opts.Repo == "" && opts.Owner == "" {
		return "", fmt.Errorf("unable to get default branch: owner and repository name are required")
	}

	if err := m.impl.getBranchData(opts); err != nil {
		return "", fmt.Errorf("fetching default branch data data: %w", err)
	}

	return opts.Branch, nil
}

// GetDefaultBranch returns the default branch of a repository
func (m *Manager) GetLatestCommit(funcs ...ooFn) (string, error) {
	opts := &Options{}
	for _, f := range funcs {
		if err := f(opts); err != nil {
			return "", err
		}
	}

	if opts.Repo == "" && opts.Owner == "" {
		return "", fmt.Errorf("unable to get default branch: owner and repository name are required")
	}

	if opts.Branch == "" {
		if err := m.impl.getBranchData(opts); err != nil {
			return "", fmt.Errorf("reading default branch: %w", err)
		}
	}

	sha, err := m.impl.getLatestCommit(opts)
	if err != nil {
		return "", fmt.Errorf("fetching commit data: %w", err)
	}

	return sha, nil
}
