package components

import (
	"context"
	"pls/proto/view"
	"strings"
	"time"

	"pls/go-lsp/lsp/defines"
)

var (
	protoKeywordCompletionItems []defines.CompletionItem

	kindKeyword = defines.CompletionItemKindKeyword
	kindModule  = defines.CompletionItemKindModule
	kindClass   = defines.CompletionItemKindClass
	kindEnum    = defines.CompletionItemKindEnum

	defaultCompletionTimeout = time.Millisecond * 500
)

func init() {
	for _, keyword := range []string{"string", "bytes", "double", "float", "int32", "int64",
		"uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool",
		"message", "enum", "service", "rpc", "optional", "repeated", "required",
		"option", "default", "syntax", "package", "import", "extend", "oneof", "map", "reserved",
	} {
		insertText := keyword
		protoKeywordCompletionItems = append(protoKeywordCompletionItems, defines.CompletionItem{
			Kind:       &kindKeyword,
			Label:      keyword,
			InsertText: &insertText,
		})
	}
}

func Completion(ctx context.Context, req *defines.CompletionParams) (*[]defines.CompletionItem, error) {
	if !view.IsProtoFile(req.TextDocument.Uri) {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(ctx, defaultCompletionTimeout)
	defer cancel()

	proto_file, err := view.ViewManager.GetFile(req.TextDocument.Uri)
	if err != nil || proto_file.Proto() == nil {
		return nil, nil
	}
	line_str := proto_file.ReadLine(int(req.Position.Line))
	wordWithDot := getWord(line_str, int(req.Position.Character-1), true)

	var res []defines.CompletionItem

	// suggest imported packages that match what user has typed so far
	//
	// word = "google"
	// suggest = [ google.protobuf , google.api ]
	for _, pkg := range GetImportedPackages(ctx, proto_file) {
		if strings.HasPrefix(*pkg.InsertText, wordWithDot) {
			res = append(res, pkg)
		}
	}

	if req.Context.TriggerKind != defines.CompletionTriggerKindTriggerCharacter {
		res = append(res, protoKeywordCompletionItems...)
		res = append(res, CompletionInThisFile(ctx, proto_file)...)
		return &res, err
	}

	packageName := strings.TrimSuffix(wordWithDot, ".")
	res = append(res, CompletionInPackage(ctx, proto_file, packageName)...)

	return &res, nil
}

func GetImportedPackages(ctx context.Context, proto_file view.ProtoFile) (res []defines.CompletionItem) {
	unique := make(map[string]struct{})
	for _, im := range proto_file.Proto().Imports() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		import_uri, err := view.GetDocumentUriFromImportPath(proto_file.URI(), im.ProtoImport.Filename)
		if err != nil {
			continue
		}

		file, err := view.ViewManager.GetFile(import_uri)
		if err != nil {
			continue
		}

		if len(file.Proto().Packages()) == 0 {
			continue
		}

		packageName := file.Proto().Packages()[0].ProtoPackage.Name

		if _, exist := unique[packageName]; exist {
			continue
		}

		unique[packageName] = struct{}{}
		res = append(res, defines.CompletionItem{
			Label:      packageName,
			Kind:       &kindModule,
			InsertText: &packageName,
		})

	}

	return res
}

func CompletionInPackage(ctx context.Context, file view.ProtoFile, packageName string) (res []defines.CompletionItem) {
	for _, im := range file.Proto().Imports() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		import_uri, err := view.GetDocumentUriFromImportPath(file.URI(), im.ProtoImport.Filename)
		if err != nil {
			continue
		}

		imported_file, err := view.ViewManager.GetFile(import_uri)
		if err != nil {
			continue
		}

		if len(imported_file.Proto().Packages()) == 0 {
			continue
		}

		importedPackage := imported_file.Proto().Packages()[0].ProtoPackage.Name
		if importedPackage == packageName {
			res = append(res, CompletionInThisFile(ctx, imported_file)...)
		}

	}
	return res
}

func CompletionInThisFile(ctx context.Context, file view.ProtoFile) (res []defines.CompletionItem) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	for _, enums := range file.Proto().Enums() {

		doc := formatHover(SymbolDefinition{
			Type: DefinitionTypeEnum,
			Enum: enums,
		})

		res = append(res, defines.CompletionItem{
			Label:      enums.Protobuf().Name,
			Kind:       &kindEnum,
			InsertText: &enums.Protobuf().Name,
			Documentation: defines.MarkupContent{
				Kind:  defines.MarkupKindMarkdown,
				Value: doc,
			},
		})
	}

	for _, message := range file.Proto().Messages() {

		doc := formatHover(SymbolDefinition{
			Type:    DefinitionTypeMessage,
			Message: message,
		})

		res = append(res, defines.CompletionItem{
			Label:      message.Protobuf().Name,
			Kind:       &kindClass,
			InsertText: &message.Protobuf().Name,
			Documentation: defines.MarkupContent{
				Kind:  defines.MarkupKindMarkdown,
				Value: doc,
			},
		})
	}
	return res
}
