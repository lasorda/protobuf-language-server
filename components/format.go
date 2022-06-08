package components

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os/exec"
	"path/filepath"
	"pls/proto/view"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func Format(ctx context.Context, req *defines.DocumentFormattingParams) (result *[]defines.TextEdit, err error) {
	if !view.IsPbHeader(req.TextDocument.Uri) {
		return nil, nil
	}
	format := exec.Command("clang-format", fmt.Sprintf("--assume-filename=%v", filepath.Base(string(req.TextDocument.Uri))))
	in, err := format.StdinPipe()
	if err != nil {
		return nil, err
	}

	proto_file, err := view.ViewManager.GetFile(req.TextDocument.Uri)
	if err != nil {
		return nil, err
	}
	data, _, _ := proto_file.Read(ctx)
	go func() {
		io.WriteString(in, string(data))
		defer in.Close()
	}()

	res, err := format.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	return &[]defines.TextEdit{
		{
			Range: defines.Range{
				Start: defines.Position{Line: 0, Character: 0},
				End:   defines.Position{Line: math.MaxInt32, Character: math.MaxInt32},
			},
			NewText: string(res),
		},
	}, nil
}
