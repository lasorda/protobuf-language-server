package components

import (
	"context"
	"errors"
	"fmt"
	"pls/proto/parser"
	"pls/proto/view"
	"regexp"
	"strings"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type SymbolDefinition struct {
	Filename  string
	Position  defines.Position
	Type      string
	Enum      parser.Enum
	Message   parser.Message
	ImportUri string
}

const (
	DefinitionTypeImport  = "import"
	DefinitionTypeMessage = "message"
	DefinitionTypeEnum    = "enum"
)

var ErrSymbolNotFound = errors.New("symbol not found")

func JumpDefine(ctx context.Context, req *defines.DefinitionParams) (result *[]defines.LocationLink, err error) {

	symbols, err := findSymbolDefinition(ctx, &req.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	locations := locationFromSymbols(symbols)

	return &locations, nil
}

func locationFromSymbols(symbols []SymbolDefinition) (result []defines.LocationLink) {

	for _, symbol := range symbols {
		switch symbol.Type {
		case DefinitionTypeImport:
			result = append(result, defines.LocationLink{
				TargetUri: defines.DocumentUri(symbol.ImportUri),
			})
		case DefinitionTypeEnum:
			proto := symbol.Enum.Protobuf()
			result = append(result, defines.LocationLink{
				TargetUri: defines.DocumentUri(proto.Position.Filename),
				TargetSelectionRange: defines.Range{
					Start: defines.Position{
						Line:      symbol.Position.Line,
						Character: symbol.Position.Character,
					},
					End: defines.Position{
						Line:      symbol.Position.Line,
						Character: symbol.Position.Character + uint(len(proto.Name)),
					},
				},
			})
		case DefinitionTypeMessage:
			proto := symbol.Message.Protobuf()
			result = append(result, defines.LocationLink{
				TargetUri: defines.DocumentUri(proto.Position.Filename),
				TargetSelectionRange: defines.Range{
					Start: defines.Position{
						Line:      symbol.Position.Line,
						Character: symbol.Position.Character,
					},
					End: defines.Position{
						Line:      symbol.Position.Line,
						Character: symbol.Position.Character + uint(len(proto.Name)),
					},
				},
			})
		}
	}

	return result
}

func findSymbolDefinition(ctx context.Context, position *defines.TextDocumentPositionParams) (result []SymbolDefinition, err error) {

	if view.IsProtoFile(position.TextDocument.Uri) {
		return JumpProtoDefine(ctx, position)
	}
	if view.IsPbHeader(position.TextDocument.Uri) {
		return JumpPbHeaderDefine(ctx, position)
	}

	return nil, ErrSymbolNotFound
}

func JumpPbHeaderDefine(ctx context.Context, req *defines.TextDocumentPositionParams) (result []SymbolDefinition, err error) {
	proto_uri := strings.ReplaceAll(string(req.TextDocument.Uri), "bazel-out/local_linux-fastbuild/genfiles/", "")
	proto_uri = strings.ReplaceAll(proto_uri, ".pb.h", ".proto")
	proto_file, err := view.ViewManager.GetFile(defines.DocumentUri(proto_uri))
	if err != nil {
		return nil, err
	}
	line := view.ViewManager.GetPbHeaderLine(req.TextDocument.Uri, int(req.Position.Line))
	word := getWord(line, int(req.Position.Character), false)
	logs.Printf("line %v, word %v", line, word)
	res, err := searchType(proto_file, word)
	// better than nothing
	if (res == nil || len(res) == 0) && strings.Contains(word, "_") {
		split_res := strings.Split(word, "_")
		if len(split_res) > 0 {
			res, err = searchType(proto_file, split_res[0])
		}
	}
	return res, err
}

func JumpProtoDefine(ctx context.Context, position *defines.TextDocumentPositionParams) (result []SymbolDefinition, err error) {
	proto_file, err := view.ViewManager.GetFile(position.TextDocument.Uri)

	if err != nil {
		return nil, err
	}
	line_str := proto_file.ReadLine(int(position.Position.Line))
	if len(line_str) < int(position.Position.Character) {
		return nil, fmt.Errorf("pos %v line_str %v", position.Position, line_str)
	}

	// dont consider single line
	if strings.HasPrefix(line_str, "import") {
		return jumpImport(ctx, position, line_str)
	}

	// type define
	package_and_word := getWord(line_str, int(position.Position.Character), true)
	pos := strings.LastIndexAny(package_and_word, ".")

	my_package := ""
	if len(proto_file.Proto().Packages()) > 0 {
		my_package = proto_file.Proto().Packages()[0].ProtoPackage.Name
	}

	var package_name, word string
	word_only := true
	if pos == -1 {
		package_name = my_package
		word = package_and_word
	} else {
		package_name, word = package_and_word[0:pos], package_and_word[pos+1:]
		word_only = false
	}

	if word_only {
		res, err := searchTypeNested(proto_file, word, int(position.Position.Line+1))
		if err == nil && len(res) > 0 {
			return res, nil
		}
	}
	if my_package == package_name {
		res, err := searchType(proto_file, word)
		if err == nil && len(res) > 0 {
			return res, nil
		}
	}

	res, err := searchImport(proto_file, package_name, my_package, word, "")
	if err == nil && len(res) > 0 {
		return res, nil
	}

	return nil, nil
}

func searchImport(proto view.ProtoFile, package_name, my_package, word, kind string) (result []SymbolDefinition, err error) {
	for _, im := range proto.Proto().Imports() {

		if kind != "" && im.ProtoImport.Kind != kind {
			continue
		}

		import_uri, err := view.GetDocumentUriFromImportPath(proto.URI(), im.ProtoImport.Filename)
		if err != nil {
			continue
		}

		import_file, err := view.ViewManager.GetFile(import_uri)
		if err != nil {
			continue
		}

		packages := import_file.Proto().Packages()
		if len(packages) > 0 {
			if qualifierReferencesPackage(package_name, packages[0].ProtoPackage.Name, my_package) {
				// same packages_name in different file
				res, err := searchType(import_file, word)
				if err == nil && len(res) > 0 {
					return res, nil
				}
			}
		}
		res, err := searchImport(import_file, package_name, my_package, word, "public")
		if res != nil && len(res) > 0 {
			return res, err
		}
	}

	return nil, nil

}

func qualifierReferencesPackage(query_pkg string, candidate_pkg string, current_pkg string) bool {
	if query_pkg == candidate_pkg { // fully qualified name
		return true
	}

	// If the current package and the candidate package are within the
	// same package, then the query need not include this package prefix.
	// Example:
	//   query_pkg = "some.dependency"
	//   candidate_pkg = "common.some.dependency"
	//   current_pkg = "common.user"

	prefix := strings.TrimSuffix(candidate_pkg, "."+query_pkg)

	return current_pkg == prefix || strings.HasPrefix(current_pkg, prefix+".")
}

func jumpImport(ctx context.Context, position *defines.TextDocumentPositionParams, line_str string) (result []SymbolDefinition, err error) {
	r, _ := regexp.Compile("\"(.+)\\/([^\\/]+)\"")
	pos := r.FindStringIndex(line_str)
	if pos == nil {
		return nil, fmt.Errorf("import match failed")
	}
	import_uri, err := view.GetDocumentUriFromImportPath(position.TextDocument.Uri, line_str[pos[0]+1:pos[1]-1])
	if err != nil {
		return nil, err
	}
	return []SymbolDefinition{{
		Type:      DefinitionTypeImport,
		ImportUri: string(import_uri),
	}}, nil
}

func searchTypeNested(proto_file view.ProtoFile, word string, line int) (result []SymbolDefinition, err error) {
	// search message
	for _, message := range proto_file.Proto().GetAllParentMessage(line) {
		if message.Protobuf().Name == word {
			message.Protobuf().Position.Filename = string(proto_file.URI())
			result = append(result, messageSymbolDefinition(proto_file, message))
		}
	}
	// search enum
	for _, enum := range proto_file.Proto().GetAllParentEnum(line) {
		if enum.Protobuf().Name == word {
			enum.Protobuf().Position.Filename = string(proto_file.URI())
			result = append(result, enumSymbolDefinition(proto_file, enum))
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrSymbolNotFound, word)
	}

	return result, nil
}

func searchType(proto_file view.ProtoFile, word string) (result []SymbolDefinition, err error) {
	// search message
	for _, message := range proto_file.Proto().Messages() {
		if message.Protobuf().Name == word {
			message.Protobuf().Position.Filename = string(proto_file.URI())
			result = append(result, messageSymbolDefinition(proto_file, message))
		}
	}
	// search enum
	for _, enum := range proto_file.Proto().Enums() {
		if enum.Protobuf().Name == word {
			enum.Protobuf().Position.Filename = string(proto_file.URI())
			result = append(result, enumSymbolDefinition(proto_file, enum))
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrSymbolNotFound, word)
	}

	return result, nil
}

func messageSymbolDefinition(proto_file view.ProtoFile, message parser.Message) SymbolDefinition {
	line := proto_file.ReadLine(message.Protobuf().Position.Line - 1)
	symbolStart := strings.Index(line, message.Protobuf().Name)
	return SymbolDefinition{
		Filename: string(proto_file.URI()),
		Position: defines.Position{
			Line:      uint(message.Protobuf().Position.Line - 1),
			Character: uint(symbolStart),
		},
		Type:    DefinitionTypeMessage,
		Message: message,
	}
}

func enumSymbolDefinition(proto_file view.ProtoFile, enum parser.Enum) SymbolDefinition {
	line := proto_file.ReadLine(enum.Protobuf().Position.Line - 1)
	symbolStart := strings.Index(line, enum.Protobuf().Name)
	return SymbolDefinition{
		Filename: string(proto_file.URI()),
		Position: defines.Position{
			Line:      uint(enum.Protobuf().Position.Line - 1),
			Character: uint(symbolStart),
		},
		Type: DefinitionTypeEnum,
		Enum: enum,
	}
}

func getWord(line string, idx int, includeDot bool) string {
	if len(line) == 0 {
		return ""
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(line) {
		idx = len(line) - 1
	}
	l, r := idx, idx

	isWordChar := func(ch byte) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || (ch == '.' && includeDot)
	}

	for l >= 0 && isWordChar(line[l]) {
		l--
	}
	if l != idx {
		l += 1
	}

	for r < len(line) && isWordChar(line[r]) {
		r++
	}
	return line[l:r]
}
