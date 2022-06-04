package types

type ProtoType string

// https://developers.google.com/protocol-buffers/docs/proto3
const (
	Double   ProtoType = "double"
	Float    ProtoType = "float"
	Int32    ProtoType = "int32"
	Int64    ProtoType = "int64"
	Uint32   ProtoType = "uint32"
	Uint64   ProtoType = "uint64"
	Sint32   ProtoType = "sint32"
	Sint64   ProtoType = "sint64"
	Fixed32  ProtoType = "fixed32"
	Fixed64  ProtoType = "fixed64"
	Sfixed32 ProtoType = "sfixed32"
	Sfixed64 ProtoType = "sfixed64"
	Bool     ProtoType = "bool"
	String   ProtoType = "string"
	Bytes    ProtoType = "bytes"
)

var BuildInProtoTypes = []ProtoType{
	Double,
	Float,
	Int32,
	Int64,
	Uint32,
	Uint64,
	Sint32,
	Sint64,
	Fixed32,
	Fixed64,
	Sfixed32,
	Sfixed64,
	Bool,
	String,
	Bytes,
}
