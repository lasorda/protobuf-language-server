
# protobuf-language-server

A language server implementation for Google Protocol Buffers

I created this tool primarily to streamline my own workflow. While some implementations might not be optimal and the features may feel incomplete, it serves my needs well enough as it is. That said, if you've got a better solution in mind, I'd be happy to switch to yours.
## installation

Build binary

```sh
go clean -modcache
# install to `go env GOPATH`
go install github.com/lasorda/protobuf-language-server@latest
```

Add it to your PATH

Configure vim/nvim

Using [coc.nvim](https://github.com/neoclide/coc.nvim), add it to `:CocConfig`

```json
    "languageserver": {
        "proto" :{
            "command": "protobuf-language-server",
            "filetypes": ["proto", "cpp"],
            "settings": {
                "additional-proto-dirs": [ ]
            }
        }
    }
```

Using [lsp-config.nvim](https://github.com/neovim/nvim-lspconfig)

```lua
-- first we need to configure our custom server
local configs = require('lspconfig.configs')
local util = require('lspconfig.util')

configs.protobuf_language_server = {
    default_config = {
        cmd = { 'path/to/protobuf-language-server' },
        filetypes = { 'proto', 'cpp' },
        root_dir = util.root_pattern('.git'),
        single_file_support = true,
        settings = {
            ["additional-proto-dirs"] = [
                -- path to additional protobuf directories
                -- "vendor",
                -- "third_party",
            ]
        },
    }
}

-- then we can continue as we do with official servers
local lspconfig = require('lspconfig')
lspconfig.protobuf_language_server.setup {
    -- your custom stuff
}
```

if you use vscode, see [vscode-extension/README.md](./vscode-extension/README.md)

## features

1. Parsing document symbols
1. Go to definition
1. Symbol definition on hover
1. Format file with clang-format
1. Code completion
1. Jump from protobuf's cpp header to proto define (only global message and enum)
