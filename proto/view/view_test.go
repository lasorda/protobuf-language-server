package view

import (
	"fmt"
	"testing"

	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
	"github.com/stretchr/testify/require"
)

func Test_view_GetDocumentUriFromImportPath(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name          string
		existingFiles []string
		cwd           defines.DocumentUri
		import_name   string
		want          defines.DocumentUri
		wantErr       error
	}{
		{
			name: "import is found when it's base directory is inside project root",
			existingFiles: []string{
				"/project-dir/api/my-service.proto",
				"/project-dir/google/protobuf/empty.proto",
			},
			cwd:         defines.DocumentUri("file:///project-dir/api/my-service.proto"),
			import_name: "google/protobuf/empty.proto",

			want:    defines.DocumentUri("file:///project-dir/google/protobuf/empty.proto"),
			wantErr: nil,
		},
		{
			name: "import is not found when it's in some sub-directory",
			existingFiles: []string{
				"/project-dir/api/my-service.proto",
				"/project-dir/protobuf-dependencies/google/protobuf/empty.proto",
			},
			cwd:         defines.DocumentUri("file:///project-dir/api/my-service.proto"),
			import_name: "google/protobuf/empty.proto",

			want:    defines.DocumentUri(""),
			wantErr: ErrNotFound,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			mockFS := &MockFS{ExistingFiles: tt.existingFiles}

			v := &view{fs: mockFS}

			got, err := v.GetDocumentUriFromImportPath(tt.cwd, tt.import_name)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
