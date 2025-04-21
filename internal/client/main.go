package client

import (
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/inveracity/voki/internal/printer"
	"github.com/pborman/indent"
	"github.com/vbauerster/mpb/v8"
)

type Client struct {
	writer io.Writer // stdout writer to enable indentation on output

	EvalContext *hcl.EvalContext
	Bar         *mpb.Progress
	Parallel    bool
	Printer     *printer.Printer
	VaultToken  *string
	VaultAddr   *string
	Steps       *[]string
}

func New() *Client {
	w := indent.New(os.Stdout, "   ")
	p := printer.New()
	return &Client{writer: w, Printer: p}
}
