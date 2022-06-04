package registry

import (
	"sync"

	protobuf "github.com/emicklei/proto"
)

// Service is a registry for protobuf service.
type Service interface {
	Protobuf() *protobuf.Service

	RPCs() []*RPC

	GetRPCByName(bool string) (*RPC, bool)

	GetRPCByLine(line int) (*RPC, bool)
}

type service struct {
	protoService *protobuf.Service

	rpcs []*RPC

	rpcNameToRPC map[string]*RPC

	lineToRPC map[int]*RPC

	mu *sync.RWMutex
}

var _ Service = (*service)(nil)

// NewService returns Service initialized by provided *protobuf.Service.
func NewService(protoService *protobuf.Service) Service {
	s := &service{
		protoService: protoService,

		rpcNameToRPC: make(map[string]*RPC),

		lineToRPC: make(map[int]*RPC),

		mu: &sync.RWMutex{},
	}

	for _, e := range protoService.Elements {
		v, ok := e.(*protobuf.RPC)
		if !ok {
			continue
		}
		r := NewRPC(v)
		s.rpcs = append(s.rpcs, r)
	}

	for _, r := range s.rpcs {
		s.rpcNameToRPC[r.ProtoRPC.Name] = r
		s.lineToRPC[r.ProtoRPC.Position.Line] = r
	}

	return s
}

// Protobuf returns *protobuf.Service.
func (s *service) Protobuf() *protobuf.Service {
	return s.protoService
}

// RPCs returns slice of RPC.
func (s *service) RPCs() (rpcs []*RPC) {
	s.mu.RLock()
	rpcs = s.rpcs
	s.mu.RUnlock()
	return
}

// GetRPCByName gets RPC by provided name.
// This ensures thread safety.
func (s *service) GetRPCByName(name string) (r *RPC, ok bool) {
	s.mu.RLock()
	r, ok = s.rpcNameToRPC[name]
	s.mu.RUnlock()
	return
}

// GetRPCByLine gets RPC by provided line.
// This ensures thread safety.
func (s *service) GetRPCByLine(line int) (r *RPC, ok bool) {
	s.mu.RLock()
	r, ok = s.lineToRPC[line]
	s.mu.RUnlock()
	return
}

// RPC is a registry for protobuf rpc.
type RPC struct {
	ProtoRPC *protobuf.RPC
}

// NewRPC returns RPC initialized by provided *protobuf.RPC.
func NewRPC(protoRPC *protobuf.RPC) *RPC {
	return &RPC{
		ProtoRPC: protoRPC,
	}
}
