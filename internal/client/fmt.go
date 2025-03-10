// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
// Adapted from https://github.com/hashicorp/nomad/blob/v1.6.1/command/fmt.go

package client

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type FormatCommand struct {
	parser   *hclparse.Parser
	hclDiags hcl.Diagnostics

	errs *multierror.Error

	Recursive   bool
	WriteFile   bool
	WriteStdout bool
	Paths       []string
}

func (f *FormatCommand) Fmt() {
	for _, path := range f.Paths {
		info, err := os.Stat(path)
		if err != nil {
			f.appendError(fmt.Errorf("No file or directory at %s", path))
			continue
		}

		if info.IsDir() {
			f.processDir(path)
		} else {
			if isVokiFile(info) {
				fp, err := os.Open(path)
				if err != nil {
					f.appendError(fmt.Errorf("Failed to open file %s: %w", path, err))
					continue
				}

				f.processFile(path, fp)

				fp.Close()
			} else {
				f.appendError(fmt.Errorf("Only .hcl files can be processed"))
				continue
			}
		}
	}
}

func (f *FormatCommand) processDir(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		f.appendError(fmt.Errorf("Failed to list directory %s", path))
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		subpath := filepath.Join(path, name)

		if entry.IsDir() {
			if f.Recursive {
				f.processDir(subpath)
			}

			continue
		}

		info, err := entry.Info()
		if err != nil {
			f.appendError(err)
			continue
		}

		if isVokiFile(info) {
			fp, err := os.Open(subpath)
			if err != nil {
				f.appendError(fmt.Errorf("Failed to open file %s: %w", path, err))
				continue
			}

			f.processFile(subpath, fp)

			fp.Close()
		}
	}
}

func (f *FormatCommand) processFile(path string, r io.Reader) {
	src, err := io.ReadAll(r)
	if err != nil {
		f.appendError(fmt.Errorf("Failed to read file %s: %w", path, err))
		return
	}

	_, syntaxDiags := hclsyntax.ParseConfig(src, path, hcl.InitialPos)
	if syntaxDiags.HasErrors() {
		f.hclDiags = append(f.hclDiags, syntaxDiags...)
		return
	}
	formattedFile, diags := hclwrite.ParseConfig(src, path, hcl.InitialPos)
	if diags.HasErrors() {
		f.hclDiags = append(f.hclDiags, diags...)
		return
	}

	out := formattedFile.Bytes()

	if !bytes.Equal(src, out) {

		if f.WriteFile {
			if err := os.WriteFile(path, out, 0644); err != nil {
				f.appendError(fmt.Errorf("Failed to write file %s: %w", path, err))
				return
			}
		}

	}

	if f.WriteStdout {
		formattedFile.WriteTo(os.Stdout)
	}
}

func isVokiFile(file fs.FileInfo) bool {
	return !file.IsDir() && (filepath.Ext(file.Name()) == ".hcl")
}

func (f *FormatCommand) appendError(err error) {
	f.errs = multierror.Append(f.errs, err)
}
