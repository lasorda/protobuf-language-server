package components

import (
	"fmt"
	"testing"
)

func Test_getWord(t *testing.T) {
	type args struct {
		line       string
		idx        int
		includeDot bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "word in middle of rpc line",
			args: args{
				// cursor is right here                            |
				line:       "rpc MethodName(SearchDashboardReq) returns (SearchDashboardResp) {",
				idx:        38,
				includeDot: false,
			},
			want: "returns",
		},
		{
			name: "type in rpc parameter",
			args: args{
				// cursor is right here           |
				line:       "rpc MethodName(SearchDashboardReq) returns (SearchDashboardResp) {",
				idx:        21,
				includeDot: false,
			},
			want: "SearchDashboardReq",
		},
		{
			name: "cursor on closing parenthesis",
			args: args{
				// cursor is right here                        |
				line:       "rpc MethodName(SearchDashboardReq) returns (SearchDashboardResp) {",
				idx:        34,
				includeDot: false,
			},
			want: "",
		},
		{
			name: "qualified name without dot",
			args: args{
				// cursor is right here                                           |
				line:       "rpc MethodName(SearchDashboardReq) returns (google.protobuf.Empty) {",
				idx:        53,
				includeDot: false,
			},
			want: "protobuf",
		},
		{
			name: "qualified name with dot incomplete",
			args: args{
				// cursor is right here                                           |
				line:       "rpc MethodName(SearchDashboardReq) returns (google.protobuf.Empty) {",
				idx:        53,
				includeDot: true,
			},
			want: "google.protobuf",
		},
		{
			name: "qualified name with dot complete",
			args: args{
				// cursor is right here                                                   |
				line:       "rpc MethodName(SearchDashboardReq) returns (google.protobuf.Empty) {",
				idx:        61,
				includeDot: true,
			},
			want: "google.protobuf.Empty",
		},
		// Additional edge case tests
		{
			name: "empty line",
			args: args{
				line:       "",
				idx:        0,
				includeDot: false,
			},
			want: "",
		},
		{
			name: "negative index clamps to 0",
			args: args{
				line:       "message Foo",
				idx:        -5,
				includeDot: false,
			},
			want: "message",
		},
		{
			name: "index past end of line",
			args: args{
				line:       "Foo",
				idx:        100,
				includeDot: false,
			},
			want: "Foo",
		},
		{
			name: "word at start of line",
			args: args{
				line:       "message Request {",
				idx:        0,
				includeDot: false,
			},
			want: "message",
		},
		{
			name: "word at end of line",
			args: args{
				line:       "  string name",
				idx:        12,
				includeDot: false,
			},
			want: "name",
		},
		{
			name: "underscore in identifier",
			args: args{
				line:       "  my_field_name = 1;",
				idx:        5,
				includeDot: false,
			},
			want: "my_field_name",
		},
		{
			name: "cursor on whitespace",
			args: args{
				line:       "message   Request",
				idx:        8,
				includeDot: false,
			},
			want: "",
		},
		{
			name: "single char word",
			args: args{
				line:       "a = 1",
				idx:        0,
				includeDot: false,
			},
			want: "a",
		},
		{
			name: "number in identifier",
			args: args{
				line:       "field1 int32 = 1",
				idx:        3,
				includeDot: false,
			},
			want: "field1",
		},
	}
	for i, tt := range tests {
		name := tt.name
		if name == "" {
			name = fmt.Sprint(i)
		}
		t.Run(name, func(t *testing.T) {
			if got := getWord(tt.args.line, tt.args.idx, tt.args.includeDot); got != tt.want {
				t.Errorf("getWord() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func Test_qualifierReferencesPackage(t *testing.T) {
	tests := []struct {
		name         string
		queryPkg     string
		candidatePkg string
		currentPkg   string
		want         bool
	}{
		{
			name:         "fully qualified name matches exactly",
			queryPkg:     "google.protobuf",
			candidatePkg: "google.protobuf",
			currentPkg:   "myapp.service",
			want:         true,
		},
		{
			name:         "same package prefix allows short reference",
			queryPkg:     "some.dependency",
			candidatePkg: "common.some.dependency",
			currentPkg:   "common.user",
			want:         true,
		},
		{
			name:         "different package prefix",
			queryPkg:     "some.dependency",
			candidatePkg: "other.some.dependency",
			currentPkg:   "common.user",
			want:         false,
		},
		{
			name:         "nested package in same hierarchy",
			queryPkg:     "models",
			candidatePkg: "myapp.service.models",
			currentPkg:   "myapp.service",
			want:         true,
		},
		{
			name:         "current package equals prefix",
			queryPkg:     "types",
			candidatePkg: "myapp.types",
			currentPkg:   "myapp",
			want:         true,
		},
		{
			name:         "empty query matches nothing",
			queryPkg:     "",
			candidatePkg: "some.package",
			currentPkg:   "other.package",
			want:         false,
		},
		{
			name:         "same package references itself",
			queryPkg:     "myapp",
			candidatePkg: "myapp",
			currentPkg:   "myapp",
			want:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qualifierReferencesPackage(tt.queryPkg, tt.candidatePkg, tt.currentPkg); got != tt.want {
				t.Errorf("qualifierReferencesPackage(%q, %q, %q) = %v, want %v",
					tt.queryPkg, tt.candidatePkg, tt.currentPkg, got, tt.want)
			}
		})
	}
}

func Test_splitPackage(t *testing.T) {
	tests := []struct {
		name         string
		package_name string
		rest         string
		ok           bool
	}{
		{
			name:         "Abc",
			package_name: "",
			rest:         "Abc",
			ok:           true,
		},
		{
			name:         ".Abc",
			package_name: ".",
			rest:         "Abc",
			ok:           true,
		},
		{
			name:         "Abc.Def",
			package_name: "",
			rest:         "Abc.Def",
			ok:           true,
		},
		{
			name:         "abc.Def",
			package_name: "abc",
			rest:         "Def",
			ok:           true,
		},
		{
			name:         ".abc.Def",
			package_name: ".abc",
			rest:         "Def",
			ok:           true,
		},
		{
			name: "",
			ok:   false,
		},
		{
			name: ".",
			ok:   false,
		},
		{
			name: ".foo.bar",
			ok:   false,
		},
		{
			name: ".foo.bar.",
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			package_name, rest, ok := splitPackage(tt.name)
			if ok != tt.ok {
				t.Errorf("splitPackage(%q) = _, _, %v, want %v",
					tt.name, ok, tt.ok)
			}
			if ok && (package_name != tt.package_name || rest != tt.rest) {
				t.Errorf("splitPackage(%q) = %q, %q, %v, want %q, %q, %v",
					tt.name, package_name, rest, ok, tt.package_name, tt.rest, tt.ok)
			}
		})
	}
}
