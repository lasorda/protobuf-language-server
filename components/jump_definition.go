package components

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lasorda/protobuf-language-server/proto/parser"
	"github.com/lasorda/protobuf-language-server/proto/view"

	"github.com/lasorda/protobuf-language-server/go-lsp/logs"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
)

type SymbolDefinition struct {
	ProtoFile view.ProtoFile
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
			tr := defines.Range{
				Start: defines.Position{
					Line:      symbol.Position.Line,
					Character: symbol.Position.Character,
				},
				End: defines.Position{
					Line:      symbol.Position.Line,
					Character: symbol.Position.Character + uint(len(proto.Name)),
				},
			}
			result = append(result, defines.LocationLink{
				TargetUri:            defines.DocumentUri(proto.Position.Filename),
				TargetSelectionRange: tr,
				TargetRange:          tr,
			})
		case DefinitionTypeMessage:
			proto := symbol.Message.Protobuf()
			tr := defines.Range{
				Start: defines.Position{
					Line:      symbol.Position.Line,
					Character: symbol.Position.Character,
				},
				End: defines.Position{
					Line:      symbol.Position.Line,
					Character: symbol.Position.Character + uint(len(proto.Name)),
				},
			}

			result = append(result, defines.LocationLink{
				TargetUri:            defines.DocumentUri(proto.Position.Filename),
				TargetSelectionRange: tr,
				TargetRange:          tr,
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
	if !view.IsProtoFile(position.TextDocument.Uri) {
		return nil, nil
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
	if len(res) == 0 && strings.Contains(word, "_") {
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
	id := getWord(line_str, int(position.Position.Character), true)
	my_package := ""
	if len(proto_file.Proto().Packages()) > 0 {
		my_package = proto_file.Proto().Packages()[0].ProtoPackage.Name
	}

	// We need to split id into the package name and the rest.
	// While each segment of the package name is conventionally lower case, protobuf does not require it.
	// So in the worst case, we may need to run quite a lot of queries.
	//
	// To ensure consistently fast operations, we do not implement that for now.
	// Instead, use a heuristic to try to resolve the symbol:
	// assume the first segment of id that starts with a capital letter is a message or an enum.
	// Everything that came before that segment is part of the package name.
	package_name, rest, ok := splitPackage(id)
	if ok && package_name != "" {
		package_name = strings.TrimPrefix(package_name, ".")
		return resolvePackageSymbol(ctx, proto_file, my_package, package_name, rest), nil
	}
	return resolveLocalSymbol(ctx, proto_file, my_package, position.Position, id), nil
}

func searchImport(proto view.ProtoFile, package_name, my_package, word, kind string) (result []SymbolDefinition, err error) {
	for _, im := range proto.Proto().Imports() {

		if kind != "" && im.ProtoImport.Kind != kind {
			continue
		}

		import_uri, err := view.ViewManager.GetDocumentUriFromImportPath(proto.URI(), im.ProtoImport.Filename)
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
		if len(res) > 0 {
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
	import_uri, err := view.ViewManager.GetDocumentUriFromImportPath(position.TextDocument.Uri, line_str[pos[0]+1:pos[1]-1])
	if err != nil {
		return nil, err
	}
	return []SymbolDefinition{{
		Type:      DefinitionTypeImport,
		ImportUri: string(import_uri),
	}}, nil
}

// searchTypeNested resolves symbol within the environment at the given line.
//
// Protobuf has the following precedence ordering:
//  1. the containing message's nested types
//  2. the containing message
//  3. the containing message's parent's nested types
//  4. the containing message's parent
//  5. the containing message's parent's parent's nested types
//  6. the containing message's parent's parent
//  7. ...
func searchTypeNested(proto_file view.ProtoFile, word string, line int) (result []SymbolDefinition, err error) {
	message, ok := proto_file.Proto().GetParentMessage(line)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSymbolNotFound, word)
	}

	for message != nil {
		result = searchTypeNested0(proto_file, word, message)
		if len(result) > 0 {
			return result, nil
		}
		message = message.GetParentMessage()
	}
	return nil, fmt.Errorf("%w: %s", ErrSymbolNotFound, word)
}

func searchTypeNested0(proto_file view.ProtoFile, word string, parentMessage parser.Message) []SymbolDefinition {
	var nested []SymbolDefinition
	if message, ok := parentMessage.GetNestedMessageByName(word); ok {
		nested = append(nested, messageSymbolDefinition(proto_file, message))
	}
	if enum, ok := parentMessage.GetNestedEnumByName(word); ok {
		nested = append(nested, enumSymbolDefinition(proto_file, enum))
	}
	if len(nested) > 0 {
		return nested
	}
	if parentMessage.Protobuf().Name == word {
		return []SymbolDefinition{messageSymbolDefinition(proto_file, parentMessage)}
	}
	return nil
}

// traverseNestedType returns A.B.C when given from=A and parts=[B C].
func traverseNestedType(from []SymbolDefinition, parts []string) []SymbolDefinition {
	if len(parts) == 0 {
		return from
	}
	first, rest := parts[0], parts[1:]
	if first == "" {
		return from
	}
	for _, def := range from {
		if def.Type != DefinitionTypeMessage {
			continue
		}
		var nested []SymbolDefinition
		if msg, ok := def.Message.GetNestedMessageByName(first); ok {
			nested = append(nested, messageSymbolDefinition(def.ProtoFile, msg))
		}
		if enum, ok := def.Message.GetNestedEnumByName(first); ok {
			nested = append(nested, enumSymbolDefinition(def.ProtoFile, enum))
		}
		if len(nested) > 0 {
			return traverseNestedType(nested, rest)
		}
	}
	return nil
}

func resolvePackageSymbol(ctx context.Context, proto_file view.ProtoFile, my_package, package_name, id string) []SymbolDefinition {
	parts := strings.Split(id, ".")
	first, rest := parts[0], parts[1:]

	if my_package == package_name {
		res, err := searchType(proto_file, first)
		if err == nil && len(res) > 0 {
			return traverseNestedType(res, rest)
		}
	}

	res, _ := searchImport(proto_file, package_name, my_package, first, "")
	return traverseNestedType(res, rest)
}

func resolveLocalSymbol(ctx context.Context, proto_file view.ProtoFile, my_package string, position defines.Position, id string) []SymbolDefinition {
	parts := strings.Split(id, ".")
	first, rest := parts[0], parts[1:]
	if first == "" {
		return nil
	}
	line := int(position.Line + 1)
	if len(rest) == 0 {
		var current []SymbolDefinition
		if message, ok := proto_file.Proto().GetMessageByLine(line); ok && message.Protobuf().Name == first {
			current = append(current, messageSymbolDefinition(proto_file, message))
		}
		if enum, ok := proto_file.Proto().GetEnumByLine(line); ok && enum.Protobuf().Name == first {
			current = append(current, enumSymbolDefinition(proto_file, enum))
		}
		if len(current) > 0 {
			return current
		}
	}
	res, err := searchTypeNested(proto_file, first, line)
	if err == nil && len(res) > 0 {
		return traverseNestedType(res, rest)
	}
	return resolvePackageSymbol(ctx, proto_file, my_package, my_package, id)
}

func splitPackage(id string) (package_name, rest string, ok bool) {
	idx := 0
	for _, part := range strings.SplitAfter(id, ".") {
		if part != "" && 'A' <= part[0] && part[0] <= 'Z' {
			// idx == 1 captures the special case of ".A",
			// where we want to return package_name = ".".
			if idx == 0 || idx == 1 {
				return id[:idx], id[idx:], true
			}
			return id[:idx-1], id[idx:], true
		}
		idx += len(part)
	}
	return "", "", false
}

func searchType(proto_file view.ProtoFile, word string) (result []SymbolDefinition, err error) {
	// search message
	for _, message := range proto_file.Proto().Messages() {
		if message.Protobuf().Name == word {
			result = append(result, messageSymbolDefinition(proto_file, message))
		}
	}
	// search enum
	for _, enum := range proto_file.Proto().Enums() {
		if enum.Protobuf().Name == word {
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
	message.Protobuf().Position.Filename = string(proto_file.URI())
	return SymbolDefinition{
		ProtoFile: proto_file,
		Filename:  string(proto_file.URI()),
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
	enum.Protobuf().Position.Filename = string(proto_file.URI())
	return SymbolDefinition{
		ProtoFile: proto_file,
		Filename:  string(proto_file.URI()),
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

	isWordChar := func(ch byte, includeDot bool) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || (ch == '.' && includeDot)
	}

	for l >= 0 && isWordChar(line[l], includeDot) {
		l--
	}
	if l != idx {
		l += 1
	}

	// Don't include dot when looking ahead, since if the cursor is at
	//   abc.def.Ghi.Xyz
	//            ^
	// we want to navigate to abc.def.Ghi, not abc.def.Ghi.Xyz.
	for r < len(line) && isWordChar(line[r], false) {
		r++
	}
	return line[l:r]
}
