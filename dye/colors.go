package dye

import (
	"fmt"
)

var (
	escapeCodeReset   = "\033[0m"
	escapeCodeRed     = "\033[31m"
	escapeCodeGreen   = "\033[32m"
	escapeCodeYellow  = "\033[33m"
	escapeCodeBlue    = "\033[34m"
	escapeCodeMagenta = "\033[35m"
	escapeCodeCyan    = "\033[36m"
	escapeCodeWhite   = "\033[37m"
)

type ColorString struct {
	text  string
	color string
	bold  bool
}

func Str(s string) *ColorString {
	return &ColorString{text: s, color: ""}
}

func Strf(format string, args ...interface{}) *ColorString {
	return &ColorString{text: fmt.Sprintf(format, args...), color: ""}
}

func (c *ColorString) Apply(code string) *ColorString {
	c.color = code
	return c
}

func (c *ColorString) Bold() *ColorString {
	c.bold = true
	return c
}

func (c *ColorString) Red() string {
	c.Apply(escapeCodeRed)
	return c.Get()
}

func (c *ColorString) Green() string {
	c.Apply(escapeCodeGreen)
	return c.Get()
}

func (c *ColorString) Yellow() string {
	c.Apply(escapeCodeYellow)
	return c.Get()
}

func (c *ColorString) Blue() string {
	c.Apply(escapeCodeBlue)
	return c.Get()
}

func (c *ColorString) Magenta() string {
	c.Apply(escapeCodeMagenta)
	return c.Get()
}

func (c *ColorString) Cyan() string {
	c.Apply(escapeCodeCyan)
	return c.Get()
}

func (c *ColorString) White() string {
	c.Apply(escapeCodeWhite)
	return c.Get()
}

func (c *ColorString) Get() string {
	prefix := ""
	if c.bold {
		prefix = "\033[1m"
	}
	return fmt.Sprintf("%s%s%s%s", prefix, c.color, c.text, escapeCodeReset)
}
