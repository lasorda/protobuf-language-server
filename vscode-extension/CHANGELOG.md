# Change Log

All notable changes to the "protobuf-language-server" extension will be documented in this file.

Check [Keep a Changelog](http://keepachangelog.com/) for recommendations on how to structure this file.

## [Unreleased]

## [0.1.8] - 2026-08-25
- 跳转定义: hover import 语句时高亮完整导入路径(而非单个路径片段)
- 修复不含 `/` 的导入路径(如 `import "types.proto";`)无法跳转定义的问题
- 修复行尾注释包含引号路径时提取错误导入路径的问题

## [0.1.7] - 2026-08-11
- `additional-proto-dirs` now expands a leading `~/` to the user's home directory

## [0.1.6] - 2026-08-10
- documentSymbol: 支持嵌套 message/enum 的展示
- 跳转定义: 支持嵌套 message/enum 的 hover 与跳转引用
- 支持 hover 嵌套 message/enum 的定义位置
- 查找引用同样适用于嵌套 message/enum

## [0.1.5] - 2026-08-06
- Added `additional-proto-dirs` setting to resolve proto imports from external directories
- `additional-proto-dirs` now accepts absolute paths
- `SettingsFromInterface` now also unwraps vscode-languageclient's nested settings format

## [0.1.4] - 2026-08-04

- 同时发布到 Open VSX,支持 Cursor / VSCodium 等 Open VSX 兼容编辑器
- 新增 GitHub Actions 自动发布工作流(打 tag 时发布到两个市场并创建 GitHub Release)

## [0.1.3] - 2026-07-27

- Bug 修复和小改进

## [0.0.8]

- Initial release