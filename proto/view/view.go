package view

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lasorda/protobuf-language-server/go-lsp/logs"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/uri"

	"github.com/lasorda/protobuf-language-server/proto/parser"
)

type view struct {

	// keep track of files by document_uri and by basename, a single file may be mapped
	// to multiple document_uris, and the same basename may map to multiple files
	filesByURI  map[defines.DocumentUri]ProtoFile
	filesByBase map[string][]ProtoFile
	fileMu      *sync.RWMutex

	openFiles  map[defines.DocumentUri]bool
	openFileMu *sync.RWMutex

	pbHeaders map[defines.DocumentUri][]string
	Server    *lsp.Server
}

func (v *view) GetFile(document_uri defines.DocumentUri) (ProtoFile, error) {
	if f, ok := v.filesByURI[document_uri]; ok {
		return f, nil
	}
	// no file load try again
	err := v.loadProtoFile(document_uri)
	if err != nil {
		return nil, err
	}
	if f, ok := v.filesByURI[document_uri]; ok {
		return f, nil
	}

	return nil, fmt.Errorf("%v not found", document_uri)
}

type Diagnositcs struct {
	Method string                           `json:"method"`
	Params defines.PublishDiagnosticsParams `json:"params"`
}

// setContent sets the file contents for a file.
func (v *view) setContent(ctx context.Context, document_uri defines.DocumentUri, data []byte) {

	v.fileMu.Lock()
	defer v.fileMu.Unlock()
	if data == nil {
		delete(v.filesByURI, document_uri)
		return
	}

	pf := &protoFile{
		File: &file{
			document_uri: document_uri,
			data:         data,
			hash:         hashContent(data),
		},
	}
	if pre, ok := v.filesByURI[document_uri]; ok {
		pf.proto = pre.Proto()
	}
	v.filesByURI[document_uri] = pf
	// TODO:
	//  Control times of parse of proto.
	//  Currently it parses every time of file change.
	proto, err := parseProto(document_uri, data)

	defer v.sendDiagnose(document_uri, err)
	if err != nil {
		return
	}
	pf.proto = proto
}

func (v *view) shutdown(ctx context.Context) error {
	// return ViewManagerInstance.RemoveView(ctx, v)
	return nil
}

func (v *view) didOpen(document_uri defines.DocumentUri, text []byte) {
	v.openFileMu.Lock()
	v.openFiles[document_uri] = true
	v.openFileMu.Unlock()
	v.openFile(document_uri, text)
	// not like include
	v.parseImportProto(document_uri)
}

func (v *view) didOpenPbHeader(document_uri defines.DocumentUri, text string) {
	v.pbHeaders[document_uri] = strings.Split(text, "\n")
}

func (v *view) GetPbHeaderLine(document_uri defines.DocumentUri, line int) string {
	lines, ok := v.pbHeaders[document_uri]
	if !ok || len(lines) <= line {
		return ""
	}

	return lines[line]
}
func (v *view) didSave(document_uri defines.DocumentUri) {
	v.fileMu.Lock()
	if file, ok := v.filesByURI[document_uri]; ok {
		file.SetSaved(true)
	}
	v.fileMu.Unlock()
}

func (v *view) didClose(document_uri defines.DocumentUri) {
	v.openFileMu.Lock()
	delete(v.openFiles, document_uri)
	v.openFileMu.Unlock()
}

func (v *view) isOpen(document_uri defines.DocumentUri) bool {
	v.openFileMu.RLock()
	defer v.openFileMu.RUnlock()

	open, ok := v.openFiles[document_uri]
	if !ok {
		return false
	}
	return open
}

func (v *view) sendDiagnose(document_uri defines.DocumentUri, err error) {
	res := Diagnositcs{
		Method: "textDocument/publishDiagnostics",
		Params: defines.PublishDiagnosticsParams{
			Uri:         document_uri,
			Diagnostics: []defines.Diagnostic{},
		},
	}
	defer func() {
		ViewManager.Server.SendMsg(res)
	}()
	if err == nil {
		return
	}
	input := err.Error()
	re := regexp.MustCompile(`<input>:(\d+):(\d+)`)
	matches := re.FindStringSubmatch(input)

	line, row := 0, 0
	if len(matches) == 3 {
		line, err = strconv.Atoi(matches[1])
		if err != nil {
			logs.Printf("Error:%v\n", err)
			return
		}
		row, err = strconv.Atoi(matches[2])
		if err != nil {
			logs.Printf("Error:%v\n", err)
			return
		}
	}
	if line == 0 || row == 0 {
		return
	}
	logs.Println(row)
	severity := defines.DiagnosticSeverityError
	res.Params.Diagnostics = append(res.Params.Diagnostics, defines.Diagnostic{
		Message:  input,
		Severity: &severity,
		Range: defines.Range{
			Start: defines.Position{
				Line:      uint(line - 1),
				Character: uint(row - 1),
			},
			End: defines.Position{
				Line:      uint(line - 1),
				Character: uint(row),
			},
		},
	})
}

func (v *view) openFile(document_uri defines.DocumentUri, data []byte) {
	v.fileMu.Lock()
	defer v.fileMu.Unlock()

	pf := &protoFile{
		File: &file{
			document_uri: document_uri,
			data:         data,
			hash:         hashContent(data),
		},
	}

	proto, err := parseProto(document_uri, data)
	defer v.sendDiagnose(document_uri, err)
	if err != nil {
		return
	}
	pf.proto = proto
	v.filesByURI[document_uri] = pf
}

func (v *view) parseImportProto(document_uri defines.DocumentUri) {
	proto_file, err := v.GetFile(document_uri)
	if err != nil {
		logs.Printf("parseImportProto GetFile err:%v", err)
		return
	}
	for _, i := range proto_file.Proto().Imports() {
		import_uri, err := GetDocumentUriFromImportPath(document_uri, i.ProtoImport.Filename)
		if err != nil {
			logs.Printf("parse import err:%v", err)
			continue
		}
		proto_file, err := v.GetFile(import_uri)
		if proto_file == nil {
			v.loadProtoFile(import_uri)
		}
	}
}

func (v *view) loadProtoFile(document_uri defines.DocumentUri) error {
	data, err := os.ReadFile(uri.URI(document_uri).Filename())

	if err != nil {
		return fmt.Errorf("read file err:%v", err)
	}
	if !utf8.Valid(data) {
		data = toUtf8(data)
	}
	v.openFile(document_uri, data)
	return nil
}

func (v *view) mapFile(document_uri defines.DocumentUri, f ProtoFile) {
	v.fileMu.Lock()

	v.filesByURI[document_uri] = f
	basename := filepath.Base(uri.URI(document_uri).Filename())
	v.filesByBase[basename] = append(v.filesByBase[basename], f)

	v.fileMu.Unlock()
}

func newView() *view {
	return &view{
		filesByURI:  make(map[defines.DocumentUri]ProtoFile),
		filesByBase: make(map[string][]ProtoFile),
		fileMu:      &sync.RWMutex{},
		openFiles:   make(map[defines.DocumentUri]bool),
		openFileMu:  &sync.RWMutex{},
		pbHeaders:   make(map[defines.DocumentUri][]string),
	}
}

var ViewManager *view

func parseProto(document_uri defines.DocumentUri, data []byte) (proto parser.Proto, err error) {
	buf := bytes.NewBuffer(data)
	proto, err = parser.ParseProto(document_uri, buf)
	if err != nil {
		logs.Printf("parseProto err %v", err)
	}
	return proto, err
}

func GetDocumentUriFromImportPath(cwd defines.DocumentUri, import_name string) (defines.DocumentUri, error) {
	pos := path.Dir(uri.URI(cwd).Filename())
	var res defines.DocumentUri
	for path.Clean(pos) != "/" {
		abs_name := path.Join(pos, import_name)
		_, err := os.Stat(abs_name)
		if err == nil {
			return defines.DocumentUri(uri.New(path.Clean(abs_name))), nil
		}
		pos = path.Join(pos, "..")
	}
	return res, fmt.Errorf("import %v not found", import_name)
}

func toUtf8(iso8859_1_buf []byte) []byte {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return []byte(string(buf))
}
func hashContent(content []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(content))
}

func didOpen(ctx context.Context, params *defines.DidOpenTextDocumentParams) error {
	if IsProtoFile(params.TextDocument.Uri) {
		document_uri := params.TextDocument.Uri
		text := []byte(params.TextDocument.Text)

		ViewManager.didOpen(document_uri, text)
		return nil
	}

	if IsPbHeader(params.TextDocument.Uri) {
		ViewManager.didOpenPbHeader(params.TextDocument.Uri, params.TextDocument.Text)
	}
	return nil
}

func didChange(ctx context.Context, params *defines.DidChangeTextDocumentParams) error {
	if !IsProtoFile(params.TextDocument.Uri) {
		return nil
	}

	if len(params.ContentChanges) < 1 {
		return jsonrpc2.NewError(jsonrpc2.InternalError, "no content changes provided")
	}

	document_uri := params.TextDocument.Uri
	text := params.ContentChanges[0].Text

	ViewManager.setContent(ctx, document_uri, []byte(text.(string)))
	return nil
}

func didClose(ctx context.Context, params *defines.DidCloseTextDocumentParams) error {
	if !IsProtoFile(params.TextDocument.Uri) {
		return nil
	}

	document_uri := params.TextDocument.Uri

	ViewManager.didClose(document_uri)
	ViewManager.setContent(ctx, document_uri, nil)

	return nil
}

func didSave(_ context.Context, params *defines.DidSaveTextDocumentParams) error {
	if !IsProtoFile(params.TextDocument.Uri) {
		return nil
	}

	document_uri := defines.DocumentUri(params.TextDocument.Uri)

	ViewManager.didSave(document_uri)

	return nil
}

func onInitialized(ctx context.Context, req *defines.InitializeParams) (err error) {
	return nil
}

func onDidChangeConfiguration(ctx context.Context, req *defines.DidChangeConfigurationParams) (err error) {
	return nil
}

func Init(server *lsp.Server) {

	ViewManager = newView()
	ViewManager.Server = server

	server.OnInitialized(onInitialized)
	server.OnDidChangeConfiguration(onDidChangeConfiguration)
	server.OnDidOpenTextDocument(didOpen)
	server.OnDidChangeTextDocument(didChange)
	server.OnDidCloseTextDocument(didClose)
	server.OnDidSaveTextDocument(didSave)
}

func IsProtoFile(document_uri defines.DocumentUri) bool {
	return strings.HasSuffix(string(document_uri), ".proto")
}

func IsPbHeader(document_uri defines.DocumentUri) bool {
	return strings.HasSuffix(string(document_uri), ".pb.h")
}
