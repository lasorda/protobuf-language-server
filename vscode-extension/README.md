# protobuf-language-server

A language server implementation for Google Protocol Buffers

## how to use

1. Get code from [https://github.com/lasorda/protobuf-language-server](https://github.com/lasorda/protobuf-language-server)
2. Build the target `pls`, add `pls` to `PATH`

## features

1. documentSymbol
2. jump to defines
3. format file with clang-format

## build vscode extension(optional for deveplop)

```shell
npm install -g vsce
npm install -g yarn
npm install
vsce package --no-yarn
```
