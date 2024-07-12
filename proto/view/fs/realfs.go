package fs

import (
	"os"
)

// Wrapps OS file methods in FS interface
type RealFS struct{}

func (r *RealFS) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
