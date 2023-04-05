module pls

go 1.16

require (
	github.com/TobiasYin/go-lsp v0.0.0-20220223105953-c4c503a4442e
	github.com/emicklei/proto v1.11.2-0.20230405093244-ad7e79baafcc
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	go.lsp.dev/jsonrpc2 v0.10.0
	go.lsp.dev/uri v0.3.0
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/TobiasYin/go-lsp v0.0.0-20220223105953-c4c503a4442e => ./go-lsp
