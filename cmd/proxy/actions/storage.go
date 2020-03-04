package actions

import (
	"fmt"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/spf13/afero"
)

// GetStorage returns storage backend based on env configuration
func GetStorage(storageType string, storageConfig *config.StorageConfig, timeout time.Duration) (storage.Backend, error) {
	const op errors.Op = "actions.GetStorage"
	switch storageType {
	case "memory":
		return mem.NewStorage()
	case "disk":
		if storageConfig.Disk == nil {
			return nil, errors.E(op, "Invalid Disk Storage Configuration")
		}
		rootLocation := storageConfig.Disk.RootPath
		s, err := fs.NewStorage(rootLocation, afero.NewOsFs())
		if err != nil {
			errStr := fmt.Sprintf("could not create new storage from os fs (%s)", err)
			return nil, errors.E(op, errStr)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("storage type %s is unknown", storageType)
	}
}
