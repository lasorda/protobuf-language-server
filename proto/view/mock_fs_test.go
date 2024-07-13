package view

type MockFS struct {
	ExistingFiles []string
}

func (m *MockFS) FileExists(path string) bool {
	return contains(m.ExistingFiles, path)
}

func contains(items []string, x string) bool {
	for _, item := range items {
		if item == x {
			return true
		}
	}
	return false
}
