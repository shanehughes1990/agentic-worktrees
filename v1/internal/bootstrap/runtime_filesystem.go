package bootstrap

import (
	"fmt"
	"os"
)

func ensureRuntimeFilesystem(config BaseConfig) error {
	directories := []string{
		config.ApplicationRootPath(),
		config.RepositoriesPath(),
		config.RepositorySourcePath(),
		config.WorktreesPath(),
		config.LogsPath(),
		config.TrackerPath(),
	}
	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("create runtime directory %q: %w", directory, err)
		}
	}
	return nil
}
