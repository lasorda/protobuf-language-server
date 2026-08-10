package view

import (
	"fmt"
	"os"
	"testing"

	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
	"github.com/stretchr/testify/require"
)

func Test_SettingsFromInterface(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   interface{}
		want    Settings
		wantErr bool
	}{
		{
			name: "flat format is parsed correctly",
			input: map[string]interface{}{
				"additional-proto-dirs": []interface{}{"/deps", "vendor"},
			},
			want: Settings{AdditionalProtoDirs: []string{"/deps", "vendor"}},
		},
		{
			name: "vscode-languageclient wrapped format is unwrapped correctly",
			input: map[string]interface{}{
				"protobuf-language-server": map[string]interface{}{
					"additional-proto-dirs": []interface{}{"/deps", "vendor"},
				},
			},
			want: Settings{AdditionalProtoDirs: []string{"/deps", "vendor"}},
		},
		{
			name: "~ in additional-proto-dirs is expanded to home directory",
			input: map[string]interface{}{
				"additional-proto-dirs": []interface{}{"~/protos"},
			},
			want: Settings{AdditionalProtoDirs: []string{homeDir + "/protos"}},
		},
		{
			name:  "empty settings returns zero value",
			input: map[string]interface{}{},
			want:  Settings{},
		},
		{
			name:    "non-map input returns error",
			input:   "not a map",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SettingsFromInterface(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, *got)
		})
	}
}

func Test_view_GetDocumentUriFromImportPath(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name          string
		existingFiles []string
		settings      Settings
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
		{
			name: "all sub-directories set via settings.additional-proto-dirs are searched for proto definitions",
			existingFiles: []string{
				"/project-dir/api/my-service.proto",
				"/project-dir/protobuf-dependencies/google/protobuf/empty.proto",
			},
			settings: Settings{
				AdditionalProtoDirs: []string{"protobuf-dependencies"},
			},
			cwd:         defines.DocumentUri("file:///project-dir/api/my-service.proto"),
			import_name: "google/protobuf/empty.proto",

			want:    defines.DocumentUri("file:///project-dir/protobuf-dependencies/google/protobuf/empty.proto"),
			wantErr: nil,
		},
		{
			name: "absolute additional-proto-dir is resolved without walking up the directory tree",
			existingFiles: []string{
				"/project-dir/api/my-service.proto",
				"/external-deps/google/protobuf/empty.proto",
			},
			settings: Settings{
				AdditionalProtoDirs: []string{"/external-deps"},
			},
			cwd:         defines.DocumentUri("file:///project-dir/api/my-service.proto"),
			import_name: "google/protobuf/empty.proto",

			want:    defines.DocumentUri("file:///external-deps/google/protobuf/empty.proto"),
			wantErr: nil,
		},
		{
			name: "import not found when absolute additional-proto-dir does not contain it",
			existingFiles: []string{
				"/project-dir/api/my-service.proto",
			},
			settings: Settings{
				AdditionalProtoDirs: []string{"/external-deps"},
			},
			cwd:         defines.DocumentUri("file:///project-dir/api/my-service.proto"),
			import_name: "google/protobuf/empty.proto",

			want:    defines.DocumentUri(""),
			wantErr: ErrNotFound,
		},
		{
			name: "absolute and relative additional-proto-dirs are both searched",
			existingFiles: []string{
				"/project-dir/api/my-service.proto",
				"/external-deps/google/protobuf/empty.proto",
				"/project-dir/vendor/other/service.proto",
			},
			settings: Settings{
				AdditionalProtoDirs: []string{"/external-deps", "vendor"},
			},
			cwd:         defines.DocumentUri("file:///project-dir/api/my-service.proto"),
			import_name: "other/service.proto",

			want:    defines.DocumentUri("file:///project-dir/vendor/other/service.proto"),
			wantErr: nil,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			mockFS := &MockFS{ExistingFiles: tt.existingFiles}

			v := &view{fs: mockFS, settings: tt.settings}

			got, err := v.GetDocumentUriFromImportPath(tt.cwd, tt.import_name)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
