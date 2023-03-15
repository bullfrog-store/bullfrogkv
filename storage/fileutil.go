package storage

import "os"

const (
	papiDefaultPerm = 0755
)

// exist returns true if a file or directory exists.
func exist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func mkdir(path string, recursively bool, perm os.FileMode) error {
	if recursively {
		return os.MkdirAll(path, perm)
	}
	return os.Mkdir(path, perm)
}

func removeAll(path string) error {
	return os.RemoveAll(path)
}
