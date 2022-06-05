package view

import (
	"context"

	"pls/proto/parser"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

// File represents a source file of any type.
type File interface {
	URI() defines.DocumentUri
	Read(ctx context.Context) ([]byte, string, error)

	Saved() bool
	// TODO: Fix appropriate function name.
	SetSaved(saved bool)
}

type ProtoFile interface {
	File
	Proto() parser.Proto
	SetProto(proto parser.Proto)
}

// file is a file for changed files.
type file struct {
	document_uri defines.DocumentUri
	data         []byte
	hash         string

	// saved is true if a file has been saved on disk.
	saved bool
}

var _ File = (*file)(nil)

type protoFile struct {
	File
	proto parser.Proto
}

var _ ProtoFile = (*protoFile)(nil)

func (f *file) URI() defines.DocumentUri {
	return f.document_uri
}

func (f *file) Read(context.Context) ([]byte, string, error) {
	return f.data, f.hash, nil
}

func (f *file) Saved() bool {
	return f.saved
}

func (f *file) SetSaved(saved bool) {
	f.saved = saved
}

func (p *protoFile) Proto() parser.Proto {
	return p.proto
}

func (p *protoFile) SetProto(proto parser.Proto) {
	p.proto = proto
}
