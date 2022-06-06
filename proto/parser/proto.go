package parser

import (
	"sync"

	"github.com/TobiasYin/go-lsp/lsp/defines"
	protobuf "github.com/emicklei/proto"
)

// Proto is a registry for protobuf proto.
type Proto interface {
	Protobuf() *protobuf.Proto

	Packages() []*Package
	Messages() []Message
	Enums() []Enum
	Services() []Service
	Imports() []*Import

	GetPackageByName(name string) (*Package, bool)
	GetMessageByName(name string) (Message, bool)
	GetEnumByName(name string) (Enum, bool)
	GetServiceByName(name string) (Service, bool)

	GetPackageByLine(line int) (*Package, bool)
	GetMessageByLine(line int) (Message, bool)
	GetEnumByLine(line int) (Enum, bool)
	GetServiceByLine(line int) (Service, bool)

	GetMessageFieldByLine(line int) (*MessageField, bool)
	GetEnumFieldByLine(line int) (*EnumField, bool)
}

type proto struct {
	protoProto *protobuf.Proto

	packages []*Package
	messages []Message
	enums    []Enum
	services []Service
	imports  []*Import

	packageNameToPackage map[string]*Package
	messageNameToMessage map[string]Message
	enumNameToEnum       map[string]Enum
	serviceNameToService map[string]Service

	lineToPackage map[int]*Package
	lineToMessage map[int]Message
	lineToEnum    map[int]Enum
	lineToService map[int]Service

	mu *sync.RWMutex
}

var _ Proto = (*proto)(nil)

// NewProto returns Proto initialized by provided *protobuf.Proto.
func NewProto(document_uri defines.DocumentUri, protoProto *protobuf.Proto) Proto {
	proto := &proto{
		protoProto: protoProto,

		packageNameToPackage: make(map[string]*Package),
		messageNameToMessage: make(map[string]Message),
		enumNameToEnum:       make(map[string]Enum),
		serviceNameToService: make(map[string]Service),

		lineToPackage: make(map[int]*Package),
		lineToMessage: make(map[int]Message),
		lineToEnum:    make(map[int]Enum),
		lineToService: make(map[int]Service),

		mu: &sync.RWMutex{},
	}

	for _, el := range protoProto.Elements {
		switch v := el.(type) {

		case *protobuf.Package:
			p := NewPackage(v)
			proto.packages = append(proto.packages, p)

		case *protobuf.Message:
			m := NewMessage(v)
			proto.messages = append(proto.messages, m)

		case *protobuf.Enum:
			e := NewEnum(v)
			proto.enums = append(proto.enums, e)

		case *protobuf.Service:
			s := NewService(v)
			proto.services = append(proto.services, s)

		case *protobuf.Import:
			i := NewImport(v)
			proto.imports = append(proto.imports, i)

		default:
		}
	}

	for _, p := range proto.packages {
		proto.packageNameToPackage[p.ProtoPackage.Name] = p
		proto.lineToPackage[p.ProtoPackage.Position.Line] = p
	}

	for _, m := range proto.messages {
		proto.messageNameToMessage[m.Protobuf().Name] = m
		proto.lineToMessage[m.Protobuf().Position.Line] = m
	}

	for _, e := range proto.enums {
		proto.enumNameToEnum[e.Protobuf().Name] = e
		proto.lineToEnum[e.Protobuf().Position.Line] = e
	}

	for _, s := range proto.services {
		proto.serviceNameToService[s.Protobuf().Name] = s
		proto.lineToService[s.Protobuf().Position.Line] = s
	}
	return proto
}

// Protobuf returns *protobuf.Proto.
func (p *proto) Protobuf() *protobuf.Proto {
	return p.protoProto
}

func (p *proto) Packages() (pkgs []*Package) {
	p.mu.RLock()
	pkgs = p.packages
	p.mu.RUnlock()
	return
}

func (p *proto) Messages() (msgs []Message) {
	p.mu.RLock()
	msgs = p.messages
	p.mu.RUnlock()
	return
}

func (p *proto) Enums() (enums []Enum) {
	p.mu.RLock()
	enums = p.enums
	p.mu.RUnlock()
	return
}

func (p *proto) Services() (svcs []Service) {
	p.mu.RLock()
	svcs = p.services
	p.mu.RUnlock()
	return
}

func (p *proto) Imports() (svcs []*Import) {
	p.mu.RLock()
	svcs = p.imports
	p.mu.RUnlock()
	return
}

// GetPackageByName gets Package by provided name.
// This ensures thread safety.
func (p *proto) GetPackageByName(name string) (pkg *Package, ok bool) {
	p.mu.RLock()
	pkg, ok = p.packageNameToPackage[name]
	p.mu.RUnlock()
	return
}

// GetMessageByName gets message by provided name.
// This ensures thread safety.
func (p *proto) GetMessageByName(name string) (m Message, ok bool) {
	p.mu.RLock()
	m, ok = p.messageNameToMessage[name]
	p.mu.RUnlock()
	return
}

// GetEnumByName gets enum by provided name.
// This ensures thread safety.
func (p *proto) GetEnumByName(name string) (e Enum, ok bool) {
	p.mu.RLock()
	e, ok = p.enumNameToEnum[name]
	p.mu.RUnlock()
	return
}

// GetServiceByName gets service by provided name.
// This ensures thread safety.
func (p *proto) GetServiceByName(name string) (s Service, ok bool) {
	p.mu.RLock()
	s, ok = p.serviceNameToService[name]
	p.mu.RUnlock()
	return
}

// GetPackageByLine gets Package by provided line.
// This ensures thread safety.
func (p *proto) GetPackageByLine(line int) (pkg *Package, ok bool) {
	p.mu.RLock()
	pkg, ok = p.lineToPackage[line]
	p.mu.RUnlock()
	return
}

// GetMessageByLine gets message by provided line.
// This ensures thread safety.
func (p *proto) GetMessageByLine(line int) (m Message, ok bool) {
	p.mu.RLock()
	m, ok = p.lineToMessage[line]
	p.mu.RUnlock()
	return
}

// GetEnumByLine gets enum by provided line.
// This ensures thread safety.
func (p *proto) GetEnumByLine(line int) (e Enum, ok bool) {
	p.mu.RLock()
	e, ok = p.lineToEnum[line]
	p.mu.RUnlock()
	return
}

// GetServiceByLine gets service by provided line.
// This ensures thread safety.
func (p *proto) GetServiceByLine(line int) (s Service, ok bool) {
	p.mu.RLock()
	s, ok = p.lineToService[line]
	p.mu.RUnlock()
	return
}

// GetMessageFieldByLine gets message field by provided line.
// This ensures thread safety.
func (p *proto) GetMessageFieldByLine(line int) (f *MessageField, ok bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, message := range p.messages {
		f, ok = message.GetFieldByLine(line)
		if ok {
			return
		}
	}
	return
}

// GetEnumFieldByLine gets enum field by provided line.
// This ensures thread safety.
func (p *proto) GetEnumFieldByLine(line int) (f *EnumField, ok bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, enum := range p.enums {
		f, ok = enum.GetFieldByLine(line)
		if ok {
			return
		}
	}
	return
}
