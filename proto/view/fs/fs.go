package fs

type FS interface {
	FileExists(path string) bool
}
