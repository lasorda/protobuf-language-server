package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"pls/components"
	"pls/proto/view"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
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
	logPath = flag.String("logs", path.Join(home, ".pls.log"), "logs file path")
	if logPath == nil || *logPath == "" {
		logger = log.New(os.Stderr, "", 0)
		return
	}
	p := *logPath
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
	server.Run()
}
