package components

import (
	"bytes"
	"context"
	"pls/proto/parser"
	"strings"
	"text/template"

	"github.com/TobiasYin/go-lsp/lsp/defines"
	"github.com/emicklei/proto"
)

var hoverTmpl *template.Template

// Parsing templates once on server start
func init() {
	hoverTmpl = template.New("hover")
	hoverTmpl = hoverTmpl.Funcs(getCustomFuncs(hoverTmpl))
	hoverTmpl = template.Must(hoverTmpl.Parse(hoverTemplate))

	messageTmpl := template.New("message")
	messageTmpl = messageTmpl.Funcs(getCustomFuncs(hoverTmpl))
	messageTmpl = template.Must(hoverTmpl.Parse(messageTemplate))

	enumTmpl := template.New("enum")
	enumTmpl = enumTmpl.Funcs(getCustomFuncs(hoverTmpl))
	enumTmpl = template.Must(hoverTmpl.Parse(enumTemplate))
}

func Hover(ctx context.Context, req *defines.HoverParams) (result *defines.Hover, err error) {

	symbols, err := findSymbolDefinition(ctx, &req.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	if len(symbols) == 0 {
		return nil, ErrSymbolNotFound
	}

	result = &defines.Hover{
		Contents: defines.MarkupContent{
			Kind:  defines.MarkupKindMarkdown,
			Value: formatHover(symbols[0]),
		},
	}

	return result, nil
}

func formatHover(symbol SymbolDefinition) string {

	var hoverData hoverData

	switch symbol.Type {
	case DefinitionTypeEnum:
		hoverData.Enum = prepareEnumData(symbol.Enum)
	case DefinitionTypeMessage:
		hoverData.Message = prepareMessageData(symbol.Message)
	default:
		return ""
	}

	buffer := bytes.NewBuffer(nil)
	err := hoverTmpl.Execute(buffer, hoverData)
	if err != nil {
		return err.Error()
	}

	return buffer.String()
}

const hoverTemplate = "```proto" + `
{{- if .Message }}
{{- templateWithIndent "message" .Message 0 }}
{{- end }}
{{- if .Enum }}
{{- templateWithIndent "enum" .Enum 0 }}
{{- end }}
` + "```"

type hoverData struct {
	Message *messageData
	Enum    *enumData
}

const enumTemplate = `{{- define "enum" }}
{{- range .Comments }}
{{ . }}
{{- end }}
enum {{ .Name }} {
	{{- range .Items }}
	{{- range .Comments }}
	{{ . }}
	{{- end }}
	{{ .Name }} = {{ .Value }}; {{ if .InlineComment }}{{ .InlineComment }}{{end}}
	{{- end }}
}
{{- end }}`

type enumData struct {
	Comments []string
	Name     string
	Items    []enumItem
}

type enumItem struct {
	Comments      []string
	Name          string
	Value         int
	InlineComment string
}

type enumFieldVisitor struct {
	proto.NoopVisitor
	visitFunc func(*proto.EnumField)
}

func (v *enumFieldVisitor) VisitEnumField(ef *proto.EnumField) {
	v.visitFunc(ef)
}

func prepareEnumData(enum parser.Enum) *enumData {

	data := enumData{
		Name:  enum.Protobuf().Name,
		Items: []enumItem{},
	}

	if enum.Protobuf().Comment != nil {
		data.Comments = formatComments(enum.Protobuf().Comment.Lines)
	}

	for _, item := range enum.Protobuf().Elements {
		item.Accept(&enumFieldVisitor{visitFunc: func(ef *proto.EnumField) {

			enumItem := enumItem{
				Name:  ef.Name,
				Value: ef.Integer,
			}

			if ef.Comment != nil {
				enumItem.Comments = formatComments(ef.Comment.Lines)
			}

			if ef.InlineComment != nil && len(ef.InlineComment.Lines) != 0 {
				enumItem.InlineComment = formatComments(ef.InlineComment.Lines[0:1])[0]
			}

			data.Items = append(data.Items, enumItem)
		}})
	}

	return &data
}

const messageTemplate = `{{- define "message" }}
{{- range .Comments }}
{{ . }}
{{- end }}
message {{ .Name }} {
{{- range .NestedEnums }}
	{{- templateWithIndent "enum" . 1 }}
{{- end }}
{{- range .NestedMessages }}
	{{- templateWithIndent "message" . 1 }}
{{- end }}
{{- range .Fields -}}
	{{- range .Comments }}
	{{ . }}
	{{- end }}
	{{.Optional}}{{.Repeated}}{{.Type}} {{.Name}} = {{.ProtoSequence}};{{ if .InlineComment }} // {{.InlineComment }}{{ end }}
{{- end }}
}
{{- end }}`

type messageData struct {
	Comments       []string
	Name           string
	Fields         []field
	NestedEnums    []*enumData
	NestedMessages []*messageData
}

type field struct {
	Comments      []string
	Repeated      string
	Optional      string
	Type          string
	Name          string
	ProtoSequence int
	InlineComment string
}

func prepareMessageData(message parser.Message) *messageData {

	data := messageData{
		Name: message.Protobuf().Name,
	}

	if message.Protobuf().Comment != nil {
		data.Comments = formatComments(message.Protobuf().Comment.Lines)
	}

	for _, nestedMsg := range message.NestedMessages() {
		data.NestedMessages = append(data.NestedMessages, prepareMessageData(nestedMsg))
	}

	for _, nestedEnum := range message.NestedEnums() {
		data.NestedEnums = append(data.NestedEnums, prepareEnumData(nestedEnum))
	}

	for _, item := range message.Fields() {

		var field field

		if item.ProtoField.Comment != nil {
			field.Comments = formatComments(item.ProtoField.Comment.Lines)
		}

		if item.ProtoField.Optional {
			field.Optional = "optional "
		}
		if item.ProtoField.Repeated {
			field.Repeated = "repeated "
		}

		field.Type = item.ProtoField.Type
		field.Name = item.ProtoField.Name
		field.ProtoSequence = item.ProtoField.Sequence

		if item.ProtoField.InlineComment != nil {
			field.InlineComment = formatComments(item.ProtoField.InlineComment.Lines[0:1])[0]
		}
		data.Fields = append(data.Fields, field)
	}

	return &data
}

func getCustomFuncs(parent *template.Template) template.FuncMap {
	return template.FuncMap{
		"templateWithIndent": func(name string, data interface{}, indent int) (string, error) {
			buf := bytes.NewBuffer(nil)
			if err := parent.ExecuteTemplate(buf, name, data); err != nil {
				return "", err
			}

			return indentText(buf.String(), indent), nil
		},
	}
}

func indentText(text string, level int) string {
	indent := strings.Repeat("\t", level)
	return indent + strings.ReplaceAll(text, "\n", "\n"+indent)
}

func formatComments(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}

	out := make([]string, 0, len(lines))
	for _, item := range lines {
		out = append(out, "//"+item)
	}
	return out
}
