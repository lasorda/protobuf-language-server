package components

import (
	"context"
	"strings"

	"github.com/lasorda/protobuf-language-server/go-lsp/logs"
	"github.com/lasorda/protobuf-language-server/go-lsp/lsp/defines"
	"github.com/lasorda/protobuf-language-server/proto/view"
)

// FindReferences finds all references to the symbol at the given position
func FindReferences(ctx context.Context, req *defines.ReferenceParams) (result *[]defines.Location, err error) {
	logs.Printf("FindReferences: uri=%s, line=%d, char=%d", req.TextDocument.Uri, req.Position.Line, req.Position.Character)

	if !view.IsProtoFile(req.TextDocument.Uri) {
		return nil, nil
	}

	// Get the proto file
	protoFile, err := view.ViewManager.GetFile(req.TextDocument.Uri)
	if err != nil {
		logs.Printf("FindReferences: GetFile error: %v", err)
		return nil, err
	}

	// Extract symbol at cursor position
	line := protoFile.ReadLine(int(req.Position.Line))
	if len(line) == 0 {
		return &[]defines.Location{}, nil
	}

	// Get the word at cursor (without package prefix for matching)
	symbolName := getWord(line, int(req.Position.Character), false)
	if symbolName == "" {
		logs.Printf("FindReferences: no symbol at position")
		return &[]defines.Location{}, nil
	}
	logs.Printf("FindReferences: looking for symbol '%s'", symbolName)

	// Find the definition to understand what type of symbol we're looking for
	symbols, _ := findSymbolDefinition(ctx, &req.TextDocumentPositionParams)

	var results []defines.Location

	// Track definition location to avoid duplicates
	var defUri defines.DocumentUri
	var defLine uint

	// If we found a definition and includeDeclaration is true, add it
	if req.Context.IncludeDeclaration && len(symbols) > 0 {
		for _, sym := range symbols {
			if sym.Type == DefinitionTypeMessage || sym.Type == DefinitionTypeEnum {
				var name string
				if sym.Type == DefinitionTypeMessage {
					name = sym.Message.Protobuf().Name
				} else {
					name = sym.Enum.Protobuf().Name
				}
				defUri = defines.DocumentUri(sym.Filename)
				defLine = sym.Position.Line
				results = append(results, defines.Location{
					Uri: defUri,
					Range: defines.Range{
						Start: sym.Position,
						End: defines.Position{
							Line:      sym.Position.Line,
							Character: sym.Position.Character + uint(len(name)),
						},
					},
				})
			}
		}
	}

	// Search for references in the current file (skip definition line to avoid duplicates)
	refs := searchFileForReferences(protoFile, symbolName, defUri, defLine)
	results = append(results, refs...)

	// Search in imported files (recursively)
	searchedFiles := make(map[defines.DocumentUri]bool)
	searchedFiles[protoFile.URI()] = true
	searchImportedFilesForReferences(protoFile, symbolName, searchedFiles, &results, defUri, defLine)

	logs.Printf("FindReferences: found %d references", len(results))
	return &results, nil
}

// searchImportedFilesForReferences recursively searches imported files for references
func searchImportedFilesForReferences(protoFile view.ProtoFile, symbolName string, searched map[defines.DocumentUri]bool, results *[]defines.Location, defUri defines.DocumentUri, defLine uint) {
	if protoFile.Proto() == nil {
		return
	}

	for _, imp := range protoFile.Proto().Imports() {
		importUri, err := view.ViewManager.GetDocumentUriFromImportPath(protoFile.URI(), imp.ProtoImport.Filename)
		if err != nil {
			continue
		}

		if searched[importUri] {
			continue
		}
		searched[importUri] = true

		importFile, err := view.ViewManager.GetFile(importUri)
		if err != nil {
			continue
		}

		refs := searchFileForReferences(importFile, symbolName, defUri, defLine)
		*results = append(*results, refs...)

		// Recursively search this file's imports
		searchImportedFilesForReferences(importFile, symbolName, searched, results, defUri, defLine)
	}
}

// searchFileForReferences searches a single file for all references to the symbol
// defUri and defLine are used to skip the definition location (to avoid duplicates)
func searchFileForReferences(protoFile view.ProtoFile, symbolName string, defUri defines.DocumentUri, defLine uint) []defines.Location {
	var results []defines.Location

	data, _, err := protoFile.Read(context.Background())
	if err != nil {
		return results
	}

	lines := strings.Split(string(data), "\n")
	fileUri := protoFile.URI()

	for lineNum, line := range lines {
		// Skip the definition line to avoid duplicates
		if fileUri == defUri && uint(lineNum) == defLine {
			continue
		}

		// Skip empty lines
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Skip import lines - we don't want to match symbol names in imports
		if strings.HasPrefix(strings.TrimSpace(line), "import") {
			continue
		}

		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Find all occurrences of the symbol in this line
		idx := 0
		for {
			foundIdx := strings.Index(line[idx:], symbolName)
			if foundIdx == -1 {
				break
			}
			foundIdx += idx // Adjust to absolute position in line

			// Verify it's a whole word (not part of a larger identifier)
			if isWholeWord(line, foundIdx, len(symbolName)) {
				results = append(results, defines.Location{
					Uri: fileUri,
					Range: defines.Range{
						Start: defines.Position{
							Line:      uint(lineNum),
							Character: uint(foundIdx),
						},
						End: defines.Position{
							Line:      uint(lineNum),
							Character: uint(foundIdx + len(symbolName)),
						},
					},
				})
			}

			idx = foundIdx + 1
			if idx >= len(line) {
				break
			}
		}
	}

	return results
}

// isWholeWord checks if the match at the given position is a complete word
func isWholeWord(line string, start int, length int) bool {
	// Check character before
	if start > 0 {
		prev := line[start-1]
		if isIdentifierChar(prev) {
			return false
		}
	}

	// Check character after
	end := start + length
	if end < len(line) {
		next := line[end]
		if isIdentifierChar(next) {
			return false
		}
	}

	return true
}

// isIdentifierChar checks if a character can be part of an identifier
func isIdentifierChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}
