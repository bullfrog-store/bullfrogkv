package storage

import "os"

// exist returns true if a file or directory exists.
func exist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
