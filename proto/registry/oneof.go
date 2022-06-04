package registry

import (
	"sync"

	protobuf "github.com/emicklei/proto"
)

// Oneof is a registry for protobuf oneof.
type Oneof interface {
	Protobuf() *protobuf.Oneof

	GetFieldByName(name string) (*OneofField, bool)

	GetFieldByLine(line int) (*OneofField, bool)
}

type oneof struct {
	protoOneofField *protobuf.Oneof

	fieldNameToField map[string]*OneofField

	lineToField map[int]*OneofField

	mu *sync.RWMutex
}

var _ Oneof = (*oneof)(nil)

// NewOneof returns Oneof initialized by provided *protobuf.Oneof.
func NewOneof(protoOneofField *protobuf.Oneof) Oneof {
	oneof := &oneof{
		protoOneofField: protoOneofField,

		fieldNameToField: make(map[string]*OneofField),

		lineToField: make(map[int]*OneofField),
	}

	for _, e := range protoOneofField.Elements {
		v, ok := e.(*protobuf.OneOfField)
		if !ok {
			continue
		}
		f := NewOneofField(v)
		oneof.fieldNameToField[v.Name] = f
		oneof.lineToField[v.Position.Line] = f
	}

	return oneof
}

// Protobuf returns *protobuf.Oneof.
func (o *oneof) Protobuf() *protobuf.Oneof {
	return o.protoOneofField
}

// GetFieldByName gets EnumField  by provided name.
// This ensures thread safety.
func (o *oneof) GetFieldByName(name string) (f *OneofField, ok bool) {
	o.mu.RLock()
	f, ok = o.fieldNameToField[name]
	o.mu.RUnlock()
	return
}

// GetFieldByName gets MapField by provided line.
// This ensures thread safety.
func (o *oneof) GetFieldByLine(line int) (f *OneofField, ok bool) {
	o.mu.RLock()
	f, ok = o.lineToField[line]
	o.mu.RUnlock()
	return
}

// OneofField is a registry for protobuf oneof field.
type OneofField struct {
	ProtoOneOfField *protobuf.OneOfField
}

// NewOneofField returns OneofField initialized by provided *protobuf.OneofField.
func NewOneofField(protoOneOfField *protobuf.OneOfField) *OneofField {
	return &OneofField{
		ProtoOneOfField: protoOneOfField,
	}
}
