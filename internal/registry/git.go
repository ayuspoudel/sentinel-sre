package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type GitAuth struct {
	// Path to SSH private key (deploy key)
	SSHKeyPath string

	// Environment variable containing HTTPS token
	TokenEnv string
}

type GitRegistry struct {
	repoURL string
	branch  string
	workdir string
	auth    GitAuth

	fs *FSRegistry
}

func NewGitRegistry(repoURL, branch, workdir string, auth GitAuth) *GitRegistry {
	return &GitRegistry{
		repoURL: repoURL,
		branch:  branch,
		workdir: workdir,
		auth:    auth,
	}
}

func (r *GitRegistry) Load(ctx context.Context) error {
	if err := os.MkdirAll(r.workdir, 0755); err != nil {
		return err
	}

	repoPath := filepath.Join(r.workdir, "repo")

	authMethod, err := r.buildAuth()
	if err != nil {
		return err
	}

	var repo *git.Repository

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// First clone
		repo, err = git.PlainCloneContext(ctx, repoPath, false, &git.CloneOptions{
			URL:           r.repoURL,
			ReferenceName: plumbing.NewBranchReferenceName(r.branch),
			SingleBranch:  true,
			Depth:         1,
			Auth:          authMethod,
		})
		if err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	} else {
		// Pull latest
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			return err
		}

		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		err = w.PullContext(ctx, &git.PullOptions{
			ReferenceName: plumbing.NewBranchReferenceName(r.branch),
			SingleBranch:  true,
			Auth:          authMethod,
		})

		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("git pull failed: %w", err)
		}
	}

	// Delegate manifest loading + validation
	r.fs = NewFSRegistry(repoPath)
	return r.fs.Load(ctx)
}

func (r *GitRegistry) Guards() []Guard {
	if r.fs == nil {
		return nil
	}
	return r.fs.Guards()
}

func (r *GitRegistry) buildAuth() (transport.AuthMethod, error) {
	// SSH deploy key
	if r.auth.SSHKeyPath != "" {
		key, err := os.ReadFile(r.auth.SSHKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}

		signer, err := ssh.NewPublicKeys("git", key, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH auth: %w", err)
		}

		return signer, nil
	}

	// HTTPS token
	if r.auth.TokenEnv != "" {
		token := os.Getenv(r.auth.TokenEnv)
		if token == "" {
			return nil, fmt.Errorf("git token env %q is not set", r.auth.TokenEnv)
		}

		return &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}, nil
	}

	// Public repo (no auth)
	return nil, nil
}
