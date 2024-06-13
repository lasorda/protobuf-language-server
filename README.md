
# pls

A language server implementation for Google Protocol Buffers

## how to use

1. build the target `pls`, add `pls` to `PATH`
2. for `coc.nvim`, `:CocConfig` like this

```json
    "languageserver": {
        "proto" :{
            "command": "pls",
            "filetypes": ["proto", "cpp"]
        }
    }
```

if you use vscode, see `vscode-extension/README.md`

## features

1. Parsing document symbols
2. Go to definition
3. Format file with clang-format
4. Code completion
5. Jump from protobuf's cpp header to proto define (only global message and enum)
