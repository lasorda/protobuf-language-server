package components

import (
	"context"
	"fmt"
	"testing"

	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
	"github.com/lasorda/protobuf-language-server/proto/parser"
	"github.com/lasorda/protobuf-language-server/proto/view"
)

// mockProtoFile implements view.ProtoFile for testing
type mockProtoFile struct {
	uri   defines.DocumentUri
	data  []byte
	proto parser.Proto
}

func (m *mockProtoFile) URI() defines.DocumentUri {
	return m.uri
}

func (m *mockProtoFile) Read(ctx context.Context) ([]byte, string, error) {
	return m.data, "", nil
}

func (m *mockProtoFile) ReadLine(line int) string {
	lines := splitLines(m.data)
	if line >= len(lines) {
		return ""
	}
	return lines[line]
}

func (m *mockProtoFile) Saved() bool {
	return true
}

func (m *mockProtoFile) SetSaved(saved bool) {}

func (m *mockProtoFile) Proto() parser.Proto {
	return m.proto
}

func (m *mockProtoFile) SetProto(proto parser.Proto) {
	m.proto = proto
}

func splitLines(data []byte) []string {
	if len(data) == 0 {
		return []string{}
	}
	result := []string{}
	start := 0
	for i, b := range data {
		if b == '\n' {
			result = append(result, string(data[start:i]))
			start = i + 1
		}
	}
	if start < len(data) {
		result = append(result, string(data[start:]))
	}
	return result
}

var _ view.ProtoFile = (*mockProtoFile)(nil)

func Test_isIdentifierChar(t *testing.T) {
	tests := []struct {
		char byte
		want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{'.', false},
		{' ', false},
		{'(', false},
		{')', false},
		{'{', false},
		{'}', false},
		{'=', false},
		{';', false},
		{'\t', false},
		{'\n', false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("char_%c", tt.char), func(t *testing.T) {
			if got := isIdentifierChar(tt.char); got != tt.want {
				t.Errorf("isIdentifierChar(%c) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func Test_isWholeWord(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		start  int
		length int
		want   bool
	}{
		{
			name:   "word at start of line",
			line:   "Message foo = 1;",
			start:  0,
			length: 7,
			want:   true,
		},
		{
			name:   "word in middle",
			line:   "  MyMessage request = 1;",
			start:  2,
			length: 9,
			want:   true,
		},
		{
			name:   "word at end of line",
			line:   "returns (Response)",
			start:  9,
			length: 8,
			want:   true,
		},
		{
			name:   "partial match at start - not whole word",
			line:   "MyMessageRequest foo = 1;",
			start:  0,
			length: 9, // "MyMessage" but line has "MyMessageRequest"
			want:   false,
		},
		{
			name:   "partial match at end - not whole word",
			line:   "RequestMessage foo = 1;",
			start:  7,
			length: 7, // "Message" but preceded by "Request"
			want:   false,
		},
		{
			name:   "word after dot (package qualifier)",
			line:   "google.protobuf.Empty",
			start:  16,
			length: 5, // "Empty"
			want:   true,
		},
		{
			name:   "word before dot",
			line:   "google.protobuf.Empty",
			start:  7,
			length: 8, // "protobuf"
			want:   true,
		},
		{
			name:   "word in parentheses",
			line:   "rpc Method(Request) returns (Response)",
			start:  11,
			length: 7, // "Request"
			want:   true,
		},
		{
			name:   "empty line",
			line:   "",
			start:  0,
			length: 0,
			want:   true,
		},
		{
			name:   "single character word",
			line:   "a = 1",
			start:  0,
			length: 1,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWholeWord(tt.line, tt.start, tt.length); got != tt.want {
				t.Errorf("isWholeWord(%q, %d, %d) = %v, want %v", tt.line, tt.start, tt.length, got, tt.want)
			}
		})
	}
}

func Test_searchFileForReferences(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		symbolName string
		defUri     defines.DocumentUri
		defLine    uint
		wantCount  int
		wantLines  []uint // expected line numbers
	}{
		{
			name: "finds multiple references",
			content: `syntax = "proto3";
message Request {
  string id = 1;
}
message Response {
  Request request = 1;
}
service MyService {
  rpc Method(Request) returns (Response);
}`,
			symbolName: "Request",
			defUri:     "",
			defLine:    0,
			wantCount:  3, // line 1 (definition), line 5 (field type), line 8 (rpc param)
			wantLines:  []uint{1, 5, 8},
		},
		{
			name: "skips definition line when specified",
			content: `syntax = "proto3";
message Request {
  string id = 1;
}
message Response {
  Request request = 1;
}`,
			symbolName: "Request",
			defUri:     "file:///test.proto",
			defLine:    1, // skip line 1 (definition)
			wantCount:  1, // only line 5 (field type)
			wantLines:  []uint{5},
		},
		{
			name: "skips import lines",
			content: `syntax = "proto3";
import "common/request.proto";
message Response {
  Request request = 1;
}`,
			symbolName: "request",
			defUri:     "",
			defLine:    0,
			wantCount:  1, // only line 3 (field name), import line is skipped
			wantLines:  []uint{3},
		},
		{
			name: "skips comment lines",
			content: `syntax = "proto3";
// Request is a message type
message Response {
  Request request = 1;
}`,
			symbolName: "Request",
			defUri:     "",
			defLine:    0,
			wantCount:  1, // only line 3, comment is skipped
			wantLines:  []uint{3},
		},
		{
			name: "does not match partial words",
			content: `syntax = "proto3";
message RequestMessage {
  string id = 1;
}
message MyRequest {
  string id = 1;
}`,
			symbolName: "Request",
			defUri:     "",
			defLine:    0,
			wantCount:  0, // "RequestMessage" and "MyRequest" should not match
			wantLines:  []uint{},
		},
		{
			name: "finds word with package qualifier",
			content: `syntax = "proto3";
message Response {
  google.protobuf.Empty empty = 1;
}`,
			symbolName: "Empty",
			defUri:     "",
			defLine:    0,
			wantCount:  1,
			wantLines:  []uint{2},
		},
		{
			name: "empty file",
			content:    "",
			symbolName: "Request",
			defUri:     "",
			defLine:    0,
			wantCount:  0,
			wantLines:  []uint{},
		},
		{
			name: "finds multiple occurrences on same line",
			content: `syntax = "proto3";
message Foo {
  map<string, Foo> children = 1;
}`,
			symbolName: "Foo",
			defUri:     "",
			defLine:    0,
			wantCount:  2, // definition and map value type
			wantLines:  []uint{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFile := &mockProtoFile{
				uri:  "file:///test.proto",
				data: []byte(tt.content),
			}

			got := searchFileForReferences(mockFile, tt.symbolName, tt.defUri, tt.defLine)

			if len(got) != tt.wantCount {
				t.Errorf("searchFileForReferences() returned %d results, want %d", len(got), tt.wantCount)
				for i, loc := range got {
					t.Logf("  result[%d]: line %d, char %d", i, loc.Range.Start.Line, loc.Range.Start.Character)
				}
				return
			}

			// Verify line numbers
			for i, wantLine := range tt.wantLines {
				if i >= len(got) {
					break
				}
				if got[i].Range.Start.Line != wantLine {
					t.Errorf("result[%d] line = %d, want %d", i, got[i].Range.Start.Line, wantLine)
				}
			}
		})
	}
}

func Test_searchFileForReferences_characterPositions(t *testing.T) {
	content := `message Request {
  string name = 1;
}
message Response {
  Request data = 1;
}`
	mockFile := &mockProtoFile{
		uri:  "file:///test.proto",
		data: []byte(content),
	}

	results := searchFileForReferences(mockFile, "Request", "", 0)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First occurrence: "message Request {" - Request starts at character 8
	if results[0].Range.Start.Character != 8 {
		t.Errorf("first result start character = %d, want 8", results[0].Range.Start.Character)
	}
	if results[0].Range.End.Character != 15 { // 8 + len("Request")
		t.Errorf("first result end character = %d, want 15", results[0].Range.End.Character)
	}

	// Second occurrence: "  Request data = 1;" - Request starts at character 2
	if results[1].Range.Start.Character != 2 {
		t.Errorf("second result start character = %d, want 2", results[1].Range.Start.Character)
	}
	if results[1].Range.End.Character != 9 { // 2 + len("Request")
		t.Errorf("second result end character = %d, want 9", results[1].Range.End.Character)
	}
}

func Test_searchFileForReferences_uri(t *testing.T) {
	content := `message Foo {}`
	uri := defines.DocumentUri("file:///path/to/my.proto")
	mockFile := &mockProtoFile{
		uri:  uri,
		data: []byte(content),
	}

	results := searchFileForReferences(mockFile, "Foo", "", 0)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Uri != uri {
		t.Errorf("result URI = %q, want %q", results[0].Uri, uri)
	}
}
