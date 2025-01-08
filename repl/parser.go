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
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "/c history" {
			return NewSummarizeHistoryCmd(replCtx)
		}
		if trimmedLine == "/c clear" {
			return NewClearConversationCommand(replCtx)
		}
		return nil
	})
}
