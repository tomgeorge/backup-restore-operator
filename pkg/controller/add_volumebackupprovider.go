package controller

import (
	"github.com/tomgeorge/backup-restore-operator/pkg/controller/volumebackupprovider"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, volumebackupprovider.Add)
}
