package components

import (
	"context"
	"fmt"
	"pls/proto/view"
	"regexp"
	"strings"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func JumpDefine(ctx context.Context, req *defines.DefinitionParams) (result *[]defines.LocationLink, err error) {
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

	// type define in this file
	word := getWordWithDot(line_str, int(req.Position.Character))
	if !strings.ContainsAny(word, ".") {
		return jumpInThisFile(ctx, req, word)
	}
	return jumpInImport(ctx, req, word)
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

func jumpInImport(ctx context.Context, req *defines.DefinitionParams, package_and_word string) (result *[]defines.LocationLink, err error) {
	proto_file, _ := view.ViewManager.GetFile(req.TextDocument.Uri)
	pos := strings.LastIndexAny(package_and_word, ".")

	if pos == -1 {
		return nil, nil
	}
	packages_name, word := package_and_word[0:pos], package_and_word[pos+1:]

	for _, im := range proto_file.Proto().Imports() {
		import_uri, err := view.GetDocumentUriFromImportPath(req.TextDocument.Uri, im.ProtoImport.Filename)
		if err != nil {
			continue
		}

		import_file, err := view.ViewManager.GetFile(import_uri)
		if err != nil {
			continue
		}

		if len(import_file.Proto().Packages()) > 0 && import_file.Proto().Packages()[0].ProtoPackage.Name == packages_name {
			// search message
			for _, message := range import_file.Proto().Messages() {
				if message.Protobuf().Name == word {
					line := import_file.ReadLine(message.Protobuf().Position.Line - 1)
					return &[]defines.LocationLink{{
						TargetUri: import_uri,
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
		}
	}

	return nil, nil
}

func jumpInThisFile(ctx context.Context, req *defines.DefinitionParams, word string) (result *[]defines.LocationLink, err error) {
	proto_file, _ := view.ViewManager.GetFile(req.TextDocument.Uri)

	// search message
	for _, message := range proto_file.Proto().Messages() {
		if message.Protobuf().Name == word {
			line := proto_file.ReadLine(message.Protobuf().Position.Line - 1)
			return &[]defines.LocationLink{{
				TargetUri: req.TextDocument.Uri,
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
	return nil, nil
}

func getWordWithDot(line string, idx int) string {
	l, r := idx, idx
	for l >= 0 {
		if (line[l] >= 'a' && line[l] <= 'z') || (line[l] >= 'A' && line[l] <= 'Z') || line[l] == '_' || line[l] == '.' {
			l--
			continue
		}
		break
	}
	for r < len(line) {
		if (line[r] >= 'a' && line[r] <= 'z') || (line[r] >= 'A' && line[r] <= 'Z') || line[r] == '_' || line[r] == '.' {
			r++
			continue
		}
		break
	}
	return line[l+1 : r]
}
