package heliumerr

import (
	"fmt"
	"strings"
	"time"
)

type Trace struct {
	Name    string
	Entered time.Time
	File    string
}

type Error struct {
	errtype, filename, msg string
	line, col              int
	trace                  []*Trace
}

func EmptyTrace() []*Trace {
	return []*Trace{}
}

func New(msg string, trace []*Trace) *Error {
	return &Error{
		msg:     msg,
		trace:   trace,
		line:    -1,
		col:     -1,
		errtype: "",
	}
}

func (e *Error) SetPos(line, col int) *Error {
	e.line = line
	e.col = col
	return e
}

func (e *Error) SetFilename(fname string) *Error {
	e.filename = fname
	return e
}

func (e *Error) SetType(errtype string) *Error {
	e.errtype = errtype
	return e
}

func (e *Error) AddTrace(entry *Trace) *Error {
	e.trace = append(e.trace, entry)
	return e
}

func (e *Error) ClearTrace() *Error {
	e.trace = []*Trace{}
	return e
}

func (e *Error) Error() string {
	var str strings.Builder

	if len(e.trace) > 0 {
		str.WriteString("\x1b[34mTrace:\x1b[0m\n")

		for i, entry := range e.trace {
			elapsed := time.Since(entry.Entered)

			connector := "├──"
			if i == len(e.trace)-1 {
				connector = "└──"
			}

			fmt.Fprintf(&str, "\x1b[90m%s\x1b[0m [\x1b[33m%d\x1b[0m] \x1b[32m%s\x1b[0m %s \x1b[90m(%s)\x1b[0m\n",
				connector, i, entry.File, entry.Name, elapsed.Round(time.Microsecond))
		}

		str.WriteString("\n")
	}

	var posStr, typeStr string = "", ""

	if e.line > 0 && e.col > 0 {
		posStr = fmt.Sprintf(" at \x1b[35m%d:%d\x1b[0m", e.line, e.col)
	}

	if e.errtype != "" {
		typeStr = fmt.Sprintf("\x1b[34m%s\x1b[0m ", e.errtype)
	}

	fmt.Fprintf(&str, "%s\x1b[91merror\x1b[0m in file \x1b[32m%s\x1b[0m%s:\n", typeStr,
		e.filename, posStr)
	fmt.Fprintf(&str, "\x1b[90m└──\x1b[0m %s\n", e.msg)

	str.WriteString("\n")

	return str.String()
}

func (e *Error) Print() {
	fmt.Print(e.Error())
}

func Wrap(e error) *Error {
	return New(e.Error(), EmptyTrace())
}
