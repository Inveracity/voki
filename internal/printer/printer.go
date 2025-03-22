package printer

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pborman/indent"
)

type Printer struct {
	Silent bool
}

func New() *Printer {
	return &Printer{}
}

var w = indent.New(os.Stdout, "   ")

func (p *Printer) Title(text string) {
	if !p.Silent {
		fmt.Println(color.BlueString("==== " + text + " ====\n"))
	}
}

func (p *Printer) Default(text string) {
	if !p.Silent {
		fmt.Println(strings.TrimSpace(text))
	}
}

func (p *Printer) Info(text string) {
	if !p.Silent {
		fmt.Fprintln(w, color.BlueString(strings.TrimSpace(text)))
	}
}

func (p *Printer) Success(text string) {
	if !p.Silent {
		fmt.Fprintln(w, color.GreenString(strings.TrimSpace(text)))
	}
}

func (p *Printer) Error(text string) {
	if !p.Silent {
		fmt.Fprintln(w, color.RedString(strings.TrimSpace(text)))
	}
}

func (p *Printer) Fatal(err error) {
	log.Fatalln(color.RedString(err.Error()))
}
