package client

import (
	"io"
	"os"

	"github.com/pborman/indent"
)

type Client struct {
	writer io.Writer
}

func New() *Client {
	w := indent.New(os.Stdout, "   ")
	return &Client{writer: w}
}
