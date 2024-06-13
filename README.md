
# pls

A language server implementation for Google Protocol Buffers

## installation

Build binary

```sh
go build -o pls .
```

Add it to your PATH

Configure vim/nvim

Using [coc.nvim](https://github.com/neoclide/coc.nvim), add it to `:CocConfig`

```json
    "languageserver": {
        "proto" :{
            "command": "pls",
            "filetypes": ["proto", "cpp"]
        }
    }
```

if you use vscode, see [vscode-extension/README.md](./vscode-extension/README.md)

## features

1. Parsing document symbols
2. Go to definition
3. Format file with clang-format
4. Code completion
5. Jump from protobuf's cpp header to proto define (only global message and enum)
