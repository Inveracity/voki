package print

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pborman/indent"
)

var w = indent.New(os.Stdout, "   ")

func Title(text string) {
	fmt.Println(color.BlueString("==== " + text + " ====\n"))
}

func Info(text string) {
	fmt.Fprintln(w, color.BlueString(strings.TrimSpace(text)))
}

func Success(text string) {
	fmt.Fprintln(w, color.GreenString(strings.TrimSpace(text)))
}

func Error(text string) {
	fmt.Fprintln(w, color.RedString(strings.TrimSpace(text)))
}

func Fatal(err error) {
	log.Fatalln(color.RedString(err.Error()))
}
