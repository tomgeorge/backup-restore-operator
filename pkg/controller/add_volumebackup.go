package controller

import (
	"github.com/tomgeorge/backup-restore-operator/pkg/controller/volumebackup"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, volumebackup.Add)
}
