package repl

import (
	"strings"
)

type CommandParser interface {
	Parse(line string) CmdIfc
}

func ParseFunc(f func(line string) CmdIfc) CommandParser {
	return &commandParserStruct{f}
}

type commandParserFunc func(line string) CmdIfc

type commandParserStruct struct {
	f commandParserFunc
}

func (c commandParserStruct) Parse(line string) CmdIfc {
	return c.f(line)
}

func NewSlashCommandParser(replCtx *ReplContext) CommandParser {
	return ParseFunc(func(line string) CmdIfc {
		tuple := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if tuple[0] != "/c" {
			return nil
		}
		cmdName := tuple[1]
		commandDefinitions := replCtx.cmdDefinitions.Commands()
		if command, ok := commandDefinitions[cmdName]; ok {
			return command
		}
		return nil
	})
}
