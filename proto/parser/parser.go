package parser

import (
	"io"

	protobuf "github.com/emicklei/proto"

	"pls/proto/registry"
)

// ParseProtos parses protobuf files from filenames and return registry.ProtoSet.
func ParseProto(r io.Reader) (registry.Proto, error) {
	parser := protobuf.NewParser(r)
	p, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return registry.NewProto(p), nil
}
