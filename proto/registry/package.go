package registry

import protobuf "github.com/emicklei/proto"

// Package is a registry for protobuf package.
type Package struct {
	ProtoPackage *protobuf.Package
}

// NewPackage returns Package initialized by provided *protobuf.Package.
func NewPackage(protoPackage *protobuf.Package) *Package {
	return &Package{
		ProtoPackage: protoPackage,
	}
}
