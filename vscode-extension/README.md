# protobuf-language-server for VSCode

[![VSCode Marketplace](https://img.shields.io/badge/VSCode-Marketplace-blue.svg)](https://marketplace.visualstudio.com/items?itemName=panzhihao.protobuf-language-server)
[![Open VSX](https://img.shields.io/badge/Open%20VSX-Registry-orange.svg)](https://open-vsx.org/extension/panzhihao/protobuf-language-server)

A [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) implementation for Google Protocol Buffers — VSCode extension.

This extension is just the LSP client; the actual language server is provided by the `protobuf-language-server` binary.

## Install

### Option 1: Install from the marketplace (recommended)

- **VSCode Marketplace**: <https://marketplace.visualstudio.com/items?itemName=panzhihao.protobuf-language-server>
- **Open VSX Registry**: <https://open-vsx.org/extension/panzhihao/protobuf-language-server>

The Open VSX build also works in Cursor / VSCodium / Windsurf and other Open VSX-compatible editors.

### Option 2: Install the vsix manually

Download the `.vsix` from [GitHub Releases](https://github.com/lasorda/protobuf-language-server/releases), then:

```sh
code   --install-extension protobuf-language-server-0.1.8.vsix
# Cursor / VSCodium
cursor --install-extension protobuf-language-server-0.1.8.vsix
codium --install-extension protobuf-language-server-0.1.8.vsix
```

## Prerequisites

The extension looks for the `protobuf-language-server` binary on your `PATH`. Install it first:

```sh
go install github.com/lasorda/protobuf-language-server@master
```

Verify:

```sh
protobuf-language-server --help
```

If the binary is not on `PATH`, or you want to point to a custom location, set it in your VSCode `settings.json`:

```jsonc
{
    "protobuf-language-server.serverPath": "/absolute/path/to/protobuf-language-server"
}
```

## Settings

| Setting                                            | Type       | Default                       | Description                                                        |
| -------------------------------------------------- | ---------- | ----------------------------- | ------------------------------------------------------------------ |
| `protobuf-language-server.serverPath`              | `string`   | `"protobuf-language-server"`  | Path to the `protobuf-language-server` binary.                     |
| `protobuf-language-server.additional-proto-dirs`   | `string[]` | `[]`                          | Extra directories searched when resolving proto imports (relative to the workspace root or absolute). |

Example:

```jsonc
{
    "protobuf-language-server.serverPath": "protobuf-language-server",
    "protobuf-language-server.additional-proto-dirs": [
        "vendor",
        "third_party"
    ]
}
```

## Features

- Document symbol (tree view with nested message / enum)
- Go to definition (supports nested message / enum)
- Find references
- Hover (shows the definition)
- Document formatting / range formatting (`clang-format`)
- Code completion (triggered by `.`)

## Build from source (for development)

Requires Node.js and yarn.

```sh
# Install packaging tools
npm install -g vsce yarn

# Install dependencies
yarn install   # or npm install

# Compile
yarn compile   # or npm run compile

# Package the vsix
vsce package --no-yarn
```

The packaged `protobuf-language-server-<version>.vsix` will appear under `vscode-extension/`. Install it as described in "Option 2" above.

### Debug

Open the `vscode-extension/` folder in VSCode and press `F5` to launch an Extension Development Host for debugging.

## Changelog

See [CHANGELOG.md](./CHANGELOG.md).

## License

MIT
