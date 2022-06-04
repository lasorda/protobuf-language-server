package registry

import (
	"sync"

	protobuf "github.com/emicklei/proto"
)

// Message is a registry for protobuf message.
type Message interface {
	Protobuf() *protobuf.Message

	NestedMessages() []Message
	NestedEnums() []Enum
	Fields() []*MessageField
	Oneofs() []Oneof
	MapFields() []*MapField

	GetNestedMessageByName(name string) (Message, bool)
	GetNestedEnumByName(name string) (Enum, bool)

	GetFieldByName(name string) (*MessageField, bool)
	GetOneofFieldByName(name string) (Oneof, bool)
	GetMapFieldByName(name string) (*MapField, bool)

	GetFieldByLine(line int) (*MessageField, bool)
	GetOneofFieldByLine(line int) (Oneof, bool)
	GetMapFieldByLine(line int) (*MapField, bool)
}

type message struct {
	protoMessage *protobuf.Message

	fullyQualifiedName string

	nestedMessages []Message
	nestedEnums    []Enum
	fields         []*MessageField
	oneofs         []Oneof
	mapFields      []*MapField

	nestedEnumNameToEnum       map[string]Enum
	nestedMessageNameToMessage map[string]Message

	fieldNameToField           map[string]*MessageField
	oneofFieldNameToOneofField map[string]Oneof
	mapFieldNameToMapField     map[string]*MapField

	lineToField      map[int]*MessageField
	lineToOneofField map[int]Oneof
	lineToMapField   map[int]*MapField

	mu *sync.RWMutex
}

var _ Message = (*message)(nil)

// NewMessage returns Message initialized by provided *protobuf.Message.
func NewMessage(protoMessage *protobuf.Message) Message {
	m := &message{
		protoMessage: protoMessage,

		fullyQualifiedName: "",

		nestedEnumNameToEnum:       make(map[string]Enum),
		nestedMessageNameToMessage: make(map[string]Message),

		fieldNameToField:           make(map[string]*MessageField),
		oneofFieldNameToOneofField: make(map[string]Oneof),
		mapFieldNameToMapField:     make(map[string]*MapField),

		lineToField:      make(map[int]*MessageField),
		lineToOneofField: make(map[int]Oneof),
		lineToMapField:   make(map[int]*MapField),

		mu: &sync.RWMutex{},
	}

	for _, e := range protoMessage.Elements {
		switch v := e.(type) {

		case *protobuf.NormalField:
			f := NewMessageField(v)
			m.fields = append(m.fields, f)

		case *protobuf.Oneof:
			f := NewOneof(v)
			m.oneofs = append(m.oneofs, f)

		case *protobuf.MapField:
			f := NewMapField(v)
			m.mapFields = append(m.mapFields, f)

		default:
		}
	}

	for _, f := range m.fields {
		m.fieldNameToField[f.ProtoField.Name] = f
		m.lineToField[f.ProtoField.Position.Line] = f
	}

	for _, f := range m.oneofs {
		m.oneofFieldNameToOneofField[f.Protobuf().Name] = f
		m.lineToOneofField[f.Protobuf().Position.Line] = f
	}

	for _, f := range m.mapFields {
		m.mapFieldNameToMapField[f.ProtoMapField.Name] = f
		m.lineToMapField[f.ProtoMapField.Position.Line] = f
	}

	return m
}

// Protobuf returns *protobuf.Proto.
func (m *message) Protobuf() *protobuf.Message {
	return m.protoMessage
}

// NestedMessages returns slice of nested Message.
func (m *message) NestedMessages() (msgs []Message) {
	m.mu.RLock()
	msgs = m.nestedMessages
	m.mu.RUnlock()
	return
}

// NestedEnums returns slice of nested Enum.
func (m *message) NestedEnums() (enums []Enum) {
	m.mu.RLock()
	enums = m.nestedEnums
	m.mu.RUnlock()
	return
}

// Fields returns slice of MessageField.
func (m *message) Fields() (fs []*MessageField) {
	m.mu.RLock()
	fs = m.fields
	m.mu.RUnlock()
	return
}

// Oneofs returns slice of Oneof.
func (m *message) Oneofs() (fs []Oneof) {
	m.mu.RLock()
	fs = m.oneofs
	m.mu.RUnlock()
	return
}

// MapFields returns slice of MapField.
func (m *message) MapFields() (fs []*MapField) {
	m.mu.RLock()
	fs = m.mapFields
	m.mu.RUnlock()
	return
}

// GetNestedMessageByName gets Message by provided name.
// This ensures thread safety.
func (m *message) GetNestedMessageByName(name string) (msg Message, ok bool) {
	m.mu.RLock()
	msg, ok = m.nestedMessageNameToMessage[name]
	m.mu.RUnlock()
	return
}

// GetNestedEnumByName gets enum by provided name.
// This ensures thread safety.
func (m *message) GetNestedEnumByName(name string) (e Enum, ok bool) {
	m.mu.RLock()
	e, ok = m.nestedEnumNameToEnum[name]
	m.mu.RUnlock()
	return
}

// GetFieldByName gets MessageField by provided name.
// This ensures thread safety.
func (m *message) GetFieldByName(name string) (f *MessageField, ok bool) {
	m.mu.RLock()
	f, ok = m.fieldNameToField[name]
	m.mu.RUnlock()
	return
}

// GetFieldByName gets oneof by provided name.
// This ensures thread safety.
func (m *message) GetOneofFieldByName(name string) (f Oneof, ok bool) {
	m.mu.RLock()
	f, ok = m.oneofFieldNameToOneofField[name]
	m.mu.RUnlock()
	return f, ok
}

// GetMapFieldByName gets MapField by provided name.
// This ensures thread safety.
func (m *message) GetMapFieldByName(name string) (f *MapField, ok bool) {
	m.mu.RLock()
	f, ok = m.mapFieldNameToMapField[name]
	m.mu.RUnlock()
	return
}

// GetFieldByLine gets MessageField by provided line.
// This ensures thread safety.
func (m *message) GetFieldByLine(line int) (f *MessageField, ok bool) {
	m.mu.RLock()
	f, ok = m.lineToField[line]
	m.mu.RUnlock()
	return
}

// GetFieldByLine gets oneof by provided line.
// This ensures thread safety.
func (m *message) GetOneofFieldByLine(line int) (f Oneof, ok bool) {
	m.mu.RLock()
	f, ok = m.lineToOneofField[line]
	m.mu.RUnlock()
	return
}

// GetMapFieldByLine gets MapField by provided line.
// This ensures thread safety.
func (m *message) GetMapFieldByLine(line int) (f *MapField, ok bool) {
	m.mu.RLock()
	f, ok = m.lineToMapField[line]
	m.mu.RUnlock()
	return
}

// MessageField is a registry for protobuf message field.
type MessageField struct {
	ProtoField *protobuf.NormalField
}

// NewMessageField returns MessageField initialized by provided *protobuf.MessageField.
func NewMessageField(protoMessage *protobuf.NormalField) *MessageField {
	return &MessageField{
		ProtoField: protoMessage,
	}
}
