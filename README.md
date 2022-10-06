
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

1. documentSymbol
2. jump to defines
3. format file with clang-format
4. jump from protobuf's cpp header to proto define (only global message and enum)
