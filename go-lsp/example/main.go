package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/lasorda/protobuf-language-server/go-lsp/logs"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
)

func strPtr(str string) *string {
	return &str
}

var (
	address *string
	logPath *string
)

func init() {
	logPath = flag.String("logs", logs.DefaultLogFilePath(), "logs file path")
	address = flag.String("listen", "", "address on which to listen for remote connections")
}

func main() {
	flag.Parse()
	logs.Init(logPath)

	config := &lsp.Options{
		CompletionProvider: &defines.CompletionOptions{
			TriggerCharacters: &[]string{"."},
		},
	}
	if *address != "" {
		config.Address = *address
		config.Network = "tcp"
	}

	server := lsp.NewServer(config)
	server.OnHover(func(ctx context.Context, req *defines.HoverParams) (result *defines.Hover, err error) {
		logs.Println("hover: ", req)
		return &defines.Hover{Contents: defines.MarkupContent{Kind: defines.MarkupKindPlainText, Value: "hello world"}}, nil
	})

	server.OnCompletion(func(ctx context.Context, req *defines.CompletionParams) (result *[]defines.CompletionItem, err error) {
		logs.Println("completion: ", req)
		d := defines.CompletionItemKindText
		return &[]defines.CompletionItem{defines.CompletionItem{
			Label:      "code",
			Kind:       &d,
			InsertText: strPtr("Hello"),
		}}, nil
	})

	server.OnDocumentFormatting(func(ctx context.Context, req *defines.DocumentFormattingParams) (result *[]defines.TextEdit, err error) {
		logs.Println("format: ", req)
		line, err := ReadFile(req.TextDocument.Uri)
		if err != nil {
			return nil, err
		}
		res := []defines.TextEdit{}
		for i, v := range line {
			r := convertParagraphs(v)
			if v != r {
				res = append(res, defines.TextEdit{
					Range: defines.Range{
						Start: defines.Position{uint(i), 0},
						End:   defines.Position{uint(i), uint(len(v) + 1)},
					},
					NewText: r,
				})
			}
		}

		return &res, nil
	})
	server.Run()
}

func ReadFile(filename defines.DocumentUri) ([]string, error) {
	enEscapeUrl, _ := url.QueryUnescape(string(filename))
	data, err := ioutil.ReadFile(enEscapeUrl[6:])
	if err != nil {
		return nil, err
	}
	content := string(data)
	line := strings.Split(content, "\n")
	return line, nil
}

// split paragraphs into sentences, and make the sentence first char uppercase and others lowercase
func convertParagraphs(paragraph string) string {
	sentences := []string{}
	for _, sentence := range strings.Split(paragraph, ".") {
		sentence = strings.TrimSpace(sentence)
		s := []string{}
		w := strings.Split(sentence, " ")
		for i, v := range w {
			if len(v) > 0 {
				if i == 0 {
					s = append(s, strings.ToUpper(v[0:1])+strings.ToLower(v[1:]))
				} else {
					s = append(s, strings.ToLower(v))
				}
			}
		}
		if len(s) != 0 {
			sentences = append(sentences, strings.Join(s, " ")+".")
		}
	}
	return strings.Join(sentences, " ")
}
