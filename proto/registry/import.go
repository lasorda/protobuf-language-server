package registry

import (
	protobuf "github.com/emicklei/proto"
)

// Import is a registry for protobuf enum field.
type Import struct {
	ProtoImport *protobuf.Import
}

// NewImport returns MapField initialized by provided *protobuf.MapField.
func NewImport(protoImport *protobuf.Import) *Import {
	return &Import{
		ProtoImport: protoImport,
	}
}
