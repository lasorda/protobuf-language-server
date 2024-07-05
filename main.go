package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/lasorda/protobuf-language-server/components"
	"github.com/lasorda/protobuf-language-server/proto/view"

	"github.com/lasorda/protobuf-language-server/go-lsp/logs"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
)

func strPtr(str string) *string {
	return &str
}

func init() {
	var logger *log.Logger
	var logPath *string
	defer func() {
		logs.Init(logger)
	}()
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	logPath = flag.String("logs", path.Join(home, ".protobuf-language-server.log"), "logs file path")
	if logPath == nil || *logPath == "" {
		logger = log.New(os.Stderr, "", 0)
		return
	}
	p := *logPath
	if _, err := os.Stat(p); err == nil {
		os.Rename(p, p+".bak")
	}
	f, err := os.Create(p)
	if err == nil {
		logger = log.New(f, "", 0)
		return
	}
	panic(fmt.Sprintf("logs init error: %v", err))
}

func main() {
	server := lsp.NewServer(&lsp.Options{CompletionProvider: &defines.CompletionOptions{
		TriggerCharacters: &[]string{"."},
	}})

	view.Init(server)
	server.OnDocumentSymbolWithSliceDocumentSymbol(components.ProvideDocumentSymbol)
	server.OnDefinition(components.JumpDefine)
	server.OnDocumentFormatting(components.Format)
	server.OnCompletion(components.Completion)
	server.OnHover(components.Hover)
	server.OnDocumentRangeFormatting(components.FormatRange)
	server.Run()
}
