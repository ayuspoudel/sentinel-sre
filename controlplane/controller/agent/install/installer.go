package install

import "context"

type Installer interface {
	Install(ctx context.Context, cfg *InstallConfig) error
}
