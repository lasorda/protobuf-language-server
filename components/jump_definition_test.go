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
		args args
		want string
	}{
		{
			args: args{
				// cursor is right here                            |
				line:       "rpc MethodName(SearchDashboardReq) returns (SearchDashboardResp) {",
				idx:        38,
				includeDot: false,
			},
			want: "returns",
		},
		{
			args: args{
				// cursor is right here           |
				line:       "rpc MethodName(SearchDashboardReq) returns (SearchDashboardResp) {",
				idx:        21,
				includeDot: false,
			},
			want: "SearchDashboardReq",
		},
		{
			args: args{
				// cursor is right here                        |
				line:       "rpc MethodName(SearchDashboardReq) returns (SearchDashboardResp) {",
				idx:        34,
				includeDot: false,
			},
			want: "",
		},
		{
			args: args{
				// cursor is right here                                           |
				line:       "rpc MethodName(SearchDashboardReq) returns (google.protobuf.Empty) {",
				idx:        53,
				includeDot: false,
			},
			want: "protobuf",
		},
		{
			args: args{
				// cursor is right here                                           |
				line:       "rpc MethodName(SearchDashboardReq) returns (google.protobuf.Empty) {",
				idx:        53,
				includeDot: true,
			},
			want: "google.protobuf.Empty",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if got := getWord(tt.args.line, tt.args.idx, tt.args.includeDot); got != tt.want {
				t.Errorf("getWord() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
