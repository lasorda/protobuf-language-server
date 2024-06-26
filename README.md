
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

Using [lsp-config.nvim](https://github.com/neovim/nvim-lspconfig)

```lua
-- first we need to configure our custom server
local configs = require('lspconfig.configs')
local util = require('lspconfig.util')

configs.pls = {
    default_config = {
        cmd = { 'path/to/pls' },
        filetypes = { 'proto', 'cpp' },
        root_fir = util.root_pattern('.git'),
        single_file_support = true,
        settings = {},
    }
}

-- then we can continue as we do with official servers
local lspconfig = require('lspconfig')
lspconfig.pls.setup {}
```

if you use vscode, see [vscode-extension/README.md](./vscode-extension/README.md)

## features

1. Parsing document symbols
1. Go to definition
1. Symbol definition on hover
1. Format file with clang-format
1. Code completion
1. Jump from protobuf's cpp header to proto define (only global message and enum)
