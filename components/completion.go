package components

import (
	"context"
	"pls/proto/view"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func Completion(ctx context.Context, req *defines.CompletionParams) (result *[]defines.CompletionItem, err error) {
	if !view.IsProtoFile(req.TextDocument.Uri) {
		return nil, nil
	}
	proto_file, err := view.ViewManager.GetFile(req.TextDocument.Uri)
	if proto_file.Proto() == nil {
		return nil, nil
	}
	line_str := proto_file.ReadLine(int(req.Position.Line))
	word := getWord(line_str, int(req.Position.Character-1), false)
	if req.Context.TriggerKind != defines.CompletionTriggerKindTriggerCharacter {
		res, err := CompletionInThisFile(proto_file)

		kindModule := defines.CompletionItemKindModule
		for _, im := range proto_file.Proto().Imports() {
			import_uri, err := view.GetDocumentUriFromImportPath(req.TextDocument.Uri, im.ProtoImport.Filename)
			if err != nil {
				continue
			}

			file, err := view.ViewManager.GetFile(import_uri)
			if err != nil {
				continue
			}

			if len(file.Proto().Packages()) > 0 {
				*res = append(*res, defines.CompletionItem{
					Label:      file.Proto().Packages()[0].ProtoPackage.Name,
					Kind:       &kindModule,
					InsertText: &file.Proto().Packages()[0].ProtoPackage.Name,
				})
			}

		}

		return res, err
	}
	for _, im := range proto_file.Proto().Imports() {
		import_uri, err := view.GetDocumentUriFromImportPath(req.TextDocument.Uri, im.ProtoImport.Filename)
		if err != nil {
			continue
		}

		file, err := view.ViewManager.GetFile(import_uri)
		if err != nil {
			continue
		}

		if len(file.Proto().Packages()) > 0 && file.Proto().Packages()[0].ProtoPackage.Name == word {
			return CompletionInThisFile(file)
		}
	}
	return nil, nil
}

func CompletionInThisFile(file view.ProtoFile) (result *[]defines.CompletionItem, err error) {
	kindEnum := defines.CompletionItemKindEnum
	res := []defines.CompletionItem{}
	for _, enums := range file.Proto().Enums() {
		res = append(res, defines.CompletionItem{
			Label:      enums.Protobuf().Name,
			Kind:       &kindEnum,
			InsertText: &enums.Protobuf().Name,
		})
	}

	kindClass := defines.CompletionItemKindClass
	for _, message := range file.Proto().Messages() {
		res = append(res, defines.CompletionItem{
			Label:      message.Protobuf().Name,
			Kind:       &kindClass,
			InsertText: &message.Protobuf().Name,
		})
	}
	return &res, nil
}
