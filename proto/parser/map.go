package parser

import protobuf "github.com/emicklei/proto"

// MapField is a registry for protobuf enum field.
type MapField struct {
	ProtoMapField *protobuf.MapField
}

// NewMapField returns MapField initialized by provided *protobuf.MapField.
func NewMapField(protoMapField *protobuf.MapField) *MapField {
	return &MapField{
		ProtoMapField: protoMapField,
	}
}
