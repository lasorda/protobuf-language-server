package components

import (
	"context"

	protobuf "github.com/emicklei/proto"
	"github.com/lasorda/protobuf-language-server/proto/view"

	"github.com/lasorda/protobuf-language-server/go-lsp/logs"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
)

// calculateEndPosition calculates the end position for a symbol based on its elements
func calculateEndPosition(startLine int, elements []protobuf.Visitee) uint {
	if len(elements) == 0 {
		return uint(startLine)
	}

	maxLine := startLine
	for _, element := range elements {
		switch v := element.(type) {
		case *protobuf.NormalField:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
		case *protobuf.MapField:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
		case *protobuf.Oneof:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
		case *protobuf.EnumField:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
		case *protobuf.Enum:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
			// Recursively check nested enum elements
			if nestedEnd := int(calculateEndPosition(v.Position.Line, v.Elements)); nestedEnd > maxLine {
				maxLine = nestedEnd
			}
		case *protobuf.Message:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
			// Recursively check nested message elements
			if nestedEnd := int(calculateEndPosition(v.Position.Line, v.Elements)); nestedEnd > maxLine {
				maxLine = nestedEnd
			}
		case *protobuf.RPC:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
		case *protobuf.Option:
			if v.Position.Line > maxLine {
				maxLine = v.Position.Line
			}
		}
	}

	return uint(maxLine)
}

// calculateMessageEndPosition calculates the end position for a message considering all nested elements
func calculateMessageEndPosition(messageProto *protobuf.Message) uint {
	if messageProto == nil {
		return 0
	}
	return calculateEndPosition(messageProto.Position.Line, messageProto.Elements)
}

// calculateEnumEndPosition calculates the end position for an enum considering all its fields
func calculateEnumEndPosition(enumProto *protobuf.Enum) uint {
	if enumProto == nil {
		return 0
	}
	return calculateEndPosition(enumProto.Position.Line, enumProto.Elements)
}

// calculateServiceEndPosition calculates the end position for a service considering all its RPCs
func calculateServiceEndPosition(serviceProto *protobuf.Service) uint {
	if serviceProto == nil {
		return 0
	}
	return calculateEndPosition(serviceProto.Position.Line, serviceProto.Elements)
}

func ProvideDocumentSymbol(ctx context.Context, req *defines.DocumentSymbolParams) (result *[]defines.DocumentSymbol, err error) {
	if !view.IsProtoFile(req.TextDocument.Uri) {
		return nil, nil
	}
	file, err := view.ViewManager.GetFile(req.TextDocument.Uri)
	res := []defines.DocumentSymbol{}
	if err != nil {
		logs.Printf("GetFile err: %v", err)
		return &res, nil
	}
	for _, pack := range file.Proto().Packages() {
		packLine := pack.ProtoPackage.Position.Line - 1
		res = append(res, defines.DocumentSymbol{
			Name: pack.ProtoPackage.Name,
			Kind: defines.SymbolKindPackage,
			SelectionRange: defines.Range{
				Start: defines.Position{Line: uint(packLine)},
				End:   defines.Position{Line: uint(packLine)},
			},
			Range: defines.Range{
				Start: defines.Position{Line: uint(packLine)},
				End:   defines.Position{Line: uint(packLine)},
			},
		})
	}
	for _, imp := range file.Proto().Imports() {
		impLine := imp.ProtoImport.Position.Line - 1
		res = append(res, defines.DocumentSymbol{
			Name: imp.ProtoImport.Filename,
			Kind: defines.SymbolKindFile,
			SelectionRange: defines.Range{
				Start: defines.Position{Line: uint(impLine)},
				End:   defines.Position{Line: uint(impLine)},
			},
			Range: defines.Range{
				Start: defines.Position{Line: uint(impLine)},
				End:   defines.Position{Line: uint(impLine)},
			},
		})

	}
	for _, enums := range file.Proto().Enums() {
		enumProto := enums.Protobuf()
		startLine := enumProto.Position.Line - 1
		endLine := int(calculateEnumEndPosition(enumProto)) - 1
		if endLine < startLine {
			endLine = startLine
		}
		res = append(res, defines.DocumentSymbol{
			Name: enumProto.Name,
			Kind: defines.SymbolKindEnum,
			SelectionRange: defines.Range{
				Start: defines.Position{Line: uint(startLine)},
				End:   defines.Position{Line: uint(startLine)},
			},
			Range: defines.Range{
				Start: defines.Position{Line: uint(startLine)},
				End:   defines.Position{Line: uint(endLine)},
			},
		})
	}
	for _, message := range file.Proto().Messages() {
		message_proto := message.Protobuf()
		startLine := message_proto.Position.Line - 1
		endLine := int(calculateMessageEndPosition(message_proto)) - 1
		if endLine < startLine {
			endLine = startLine
		}
		res = append(res, defines.DocumentSymbol{
			Name: message_proto.Name,
			Kind: defines.SymbolKindClass,
			SelectionRange: defines.Range{
				Start: defines.Position{Line: uint(startLine)},
				End:   defines.Position{Line: uint(startLine)},
			},
			Range: defines.Range{
				Start: defines.Position{Line: uint(startLine)},
				End:   defines.Position{Line: uint(endLine)},
			},
		})
	}
	for _, service := range file.Proto().Services() {
		serviceProto := service.Protobuf()
		startLine := serviceProto.Position.Line - 1
		endLine := int(calculateServiceEndPosition(serviceProto)) - 1
		if endLine < startLine {
			endLine = startLine
		}
		service_sym := defines.DocumentSymbol{
			Name: serviceProto.Name,
			Kind: defines.SymbolKindNamespace,
			SelectionRange: defines.Range{
				Start: defines.Position{Line: uint(startLine)},
				End:   defines.Position{Line: uint(startLine)},
			},
			Range: defines.Range{
				Start: defines.Position{Line: uint(startLine)},
				End:   defines.Position{Line: uint(endLine)},
			},

			Children: &[]defines.DocumentSymbol{},
		}
		child := []defines.DocumentSymbol{}
		for _, rpc := range service.RPCs() {
			rpcLine := rpc.ProtoRPC.Position.Line - 1
			rpc := defines.DocumentSymbol{
				Name: rpc.ProtoRPC.Name,
				Kind: defines.SymbolKindMethod,
				SelectionRange: defines.Range{
					Start: defines.Position{Line: uint(rpcLine)},
					End:   defines.Position{Line: uint(rpcLine)},
				},
				Range: defines.Range{
					Start: defines.Position{Line: uint(rpcLine)},
					End:   defines.Position{Line: uint(rpcLine)},
				},
			}
			child = append(child, rpc)
		}
		service_sym.Children = &child
		res = append(res, service_sym)
	}
	// for i := 0; i < len(res); i++ {
	//     res[i].Range = res[i].SelectionRange
	// }
	// logs.Printf("dddddddd %+v", res)
	return &res, nil
}
