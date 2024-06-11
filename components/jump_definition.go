package components

import (
	"context"
	"fmt"
	"pls/proto/view"
	"regexp"
	"strings"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func JumpDefine(ctx context.Context, req *defines.DefinitionParams) (result *[]defines.LocationLink, err error) {
	if view.IsProtoFile(req.TextDocument.Uri) {
		return JumpProtoDefine(ctx, req)
	}

	if view.IsPbHeader(req.TextDocument.Uri) {
		return JumpPbHeaderDefine(ctx, req)
	}
	return nil, nil
}

func JumpPbHeaderDefine(ctx context.Context, req *defines.DefinitionParams) (result *[]defines.LocationLink, err error) {
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
	if (res == nil || len(*res) == 0) && strings.Contains(word, "_") {
		split_res := strings.Split(word, "_")
		if len(split_res) > 0 {
			res, err = searchType(proto_file, split_res[0])
		}
	}
	return res, err
}

func JumpProtoDefine(ctx context.Context, req *defines.DefinitionParams) (result *[]defines.LocationLink, err error) {
	proto_file, err := view.ViewManager.GetFile(req.TextDocument.Uri)

	if err != nil {
		return nil, err
	}
	line_str := proto_file.ReadLine(int(req.Position.Line))
	if len(line_str) < int(req.Position.Character) {
		return nil, fmt.Errorf("pos %v line_str %v", req.Position, line_str)
	}

	// dont consider single line
	if strings.HasPrefix(line_str, "import") {
		return jumpImport(ctx, req, line_str)
	}

	// type define
	package_and_word := getWord(line_str, int(req.Position.Character), true)
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
		res, err := searchTypeNested(proto_file, word, int(req.Position.Line+1))
		if err == nil && len(*res) > 0 {
			return res, nil
		}
	}
	if my_package == package_name {
		res, err := searchType(proto_file, word)
		if err == nil && len(*res) > 0 {
			return res, nil
		}
	}

	for _, im := range proto_file.Proto().Imports() {
		import_uri, err := view.GetDocumentUriFromImportPath(proto_file.URI(), im.ProtoImport.Filename)
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
				if err == nil && len(*res) > 0 {
					return res, nil
				}
			}
		}
		res, err := searchPublicImport(import_file, package_name, my_package, word)
		if res != nil && len(*res) > 0 {
			return res, err
		}
	}

	return nil, nil
}

func searchPublicImport(import_file view.ProtoFile, package_name string, my_package string, word string) (result *[]defines.LocationLink, err error) {
	for _, imp := range import_file.Proto().Imports() {
		if imp.ProtoImport.Kind == "public" {
			import_uri, err := view.GetDocumentUriFromImportPath(import_file.URI(), imp.ProtoImport.Filename)
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
					if err == nil && len(*res) > 0 {
						return res, nil
					}
				}
			}
			res, err := searchPublicImport(import_file, package_name, my_package, word)
			if res != nil && len(*res) > 0 {
				return res, err
			}
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

func jumpImport(ctx context.Context, req *defines.DefinitionParams, line_str string) (result *[]defines.LocationLink, err error) {
	r, _ := regexp.Compile("\"(.+)\\/([^\\/]+)\"")
	pos := r.FindStringIndex(line_str)
	if pos == nil {
		return nil, fmt.Errorf("import match failed")
	}
	import_uri, err := view.GetDocumentUriFromImportPath(req.TextDocument.Uri, line_str[pos[0]+1:pos[1]-1])
	if err != nil {
		return nil, err
	}
	return &[]defines.LocationLink{{
		TargetUri: import_uri,
	}}, nil
}

func searchTypeNested(proto_file view.ProtoFile, word string, line int) (result *[]defines.LocationLink, err error) {
	// search message
	for _, message := range proto_file.Proto().GetAllParentMessage(line) {
		if message.Protobuf().Name == word {
			line := proto_file.ReadLine(message.Protobuf().Position.Line - 1)
			return &[]defines.LocationLink{{
				TargetUri: proto_file.URI(),
				TargetSelectionRange: defines.Range{
					Start: defines.Position{
						Line:      uint(message.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word)),
					},
					End: defines.Position{
						Line:      uint(message.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word) + len(word)),
					}},
			}}, nil
		}
	}
	// search enum
	for _, enum := range proto_file.Proto().GetAllParentEnum(line) {
		if enum.Protobuf().Name == word {
			line := proto_file.ReadLine(enum.Protobuf().Position.Line - 1)
			return &[]defines.LocationLink{{
				TargetUri: proto_file.URI(),
				TargetSelectionRange: defines.Range{
					Start: defines.Position{
						Line:      uint(enum.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word)),
					},
					End: defines.Position{
						Line:      uint(enum.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word) + len(word)),
					}},
			}}, nil
		}
	}
	return nil, fmt.Errorf("%v not found", word)
}

func searchType(proto_file view.ProtoFile, word string) (result *[]defines.LocationLink, err error) {
	// search message
	for _, message := range proto_file.Proto().Messages() {
		if message.Protobuf().Name == word {
			line := proto_file.ReadLine(message.Protobuf().Position.Line - 1)
			return &[]defines.LocationLink{{
				TargetUri: proto_file.URI(),
				TargetSelectionRange: defines.Range{
					Start: defines.Position{
						Line:      uint(message.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word)),
					},
					End: defines.Position{
						Line:      uint(message.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word) + len(word)),
					}},
			}}, nil
		}
	}
	// search enum
	for _, enum := range proto_file.Proto().Enums() {
		if enum.Protobuf().Name == word {
			line := proto_file.ReadLine(enum.Protobuf().Position.Line - 1)
			return &[]defines.LocationLink{{
				TargetUri: proto_file.URI(),
				TargetSelectionRange: defines.Range{
					Start: defines.Position{
						Line:      uint(enum.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word)),
					},
					End: defines.Position{
						Line:      uint(enum.Protobuf().Position.Line) - 1,
						Character: uint(strings.Index(line, word) + len(word)),
					}},
			}}, nil
		}
	}
	return nil, fmt.Errorf("%v not found", word)
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
	ll := l
	for ll >= 0 {
		if !isWordChar(line[ll]) {
			break
		}
		ll--
	}
	if ll != l {
		ll = ll + 1
	}
	l = ll
	for r < len(line) {
		if !isWordChar(line[r]) {
			break
		}
		r++
	}
	return line[l:r]
}
