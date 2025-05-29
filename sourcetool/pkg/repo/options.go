package repo

type Options struct {
	Repo   string
	Owner  string
	Branch string
	// Organization to look for slsa and user forks
	UserForkOrg string
	Enforce     bool
	UseSSH      bool
	UpdateRepo  bool
}

type ooFn func(*Options) error

func WithRepo(repo string) ooFn {
	return func(o *Options) error {
		// TODO(puerco): Validate repo string
		o.Repo = repo
		return nil
	}
}

func WithOwner(repo string) ooFn {
	return func(o *Options) error {
		// TODO(puerco): Validate org string
		o.Owner = repo
		return nil
	}
}

func WithBranch(branch string) ooFn {
	return func(o *Options) error {
		o.Branch = branch
		return nil
	}
}

func WithEnforce(enforce bool) ooFn {
	return func(o *Options) error {
		o.Enforce = enforce
		return nil
	}
}

func WithUserForkOrg(org string) ooFn {
	return func(o *Options) error {
		o.UserForkOrg = org
		return nil
	}
}
