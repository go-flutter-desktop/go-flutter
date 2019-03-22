package execpath

import (
	"os"
	"path/filepath"
)

var execPath string

// ExecPath returns the absolute path for the currently running executable. The
// path is cached after first call.
func ExecPath() (string, error) {
	if execPath != "" {
		return execPath, nil
	}
	var err error
	execPath, err = os.Executable()
	if err != nil {
		return "", err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", err
	}
	return execPath, nil
}
