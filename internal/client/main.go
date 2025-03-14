package client

import (
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/pborman/indent"
)

type Client struct {
	writer io.Writer // stdout writer to enable indentation on output

	EvalContext *hcl.EvalContext
}

func New() *Client {
	w := indent.New(os.Stdout, "   ")
	return &Client{writer: w}
}
