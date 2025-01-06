package repl

import (
	"context"
	"fmt"
	"github.com/ergochat/readline"
	"github.com/go-errors/errors"
	"io"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
	"strings"
)

const MultiLinePrefix = "..."

type ReplContext struct {
	stopRepl    bool
	inputBuffer *strings.Builder
	modelName   string
	client      llm.ModelIfc
	session     llm.Conversation
	readline    *readline.Instance
	completer   readline.AutoCompleter
}

func (replCtx *ReplContext) ParseLine() CmdIfc {
	line, err := replCtx.readline.Readline()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return NewQuitCmd(replCtx)
		} else if err != nil {
			return NewPrintErrCmd(replCtx, err)
		}
	}
	if len(line) == 0 {
		return NewNoOpCmd()
	}

	if replCtx.inputBuffer.Len() == 0 {
		// If this is the first line, handle "/c" commands
		if strings.TrimSpace(line) == "/c history" {
			return NewSummarizeHistoryCmd(replCtx)
		}

		// If this is the first line and there is no multi-line prefix, then submit the input
		if !strings.HasPrefix(line, MultiLinePrefix) {
			return NewChainCmd(
				NewAppendCmd(replCtx, line),
				NewSubmitCmd(replCtx),
			)
		}

		// Otherwise, the multi-line prefix was specified, enter multi-line mode before appending the input
		return NewChainCmd(
			NewEnterMultiLineModeCmd(replCtx),
			NewAppendCmd(replCtx, strings.TrimPrefix(line, MultiLinePrefix)),
		)
	}
	// If this is not the first line and we see the submit input command ("."), then submit the input
	if line == "." {
		return NewSubmitCmd(replCtx)
	}
	return NewAppendCmd(replCtx, line)
}

func (replCtx *ReplContext) ResetInput() error {
	replCtx.inputBuffer.Reset()
	replCtx.readline.SetPrompt(prompts.FirstLine)
	if replCtx.completer != nil {
		newConfig := replCtx.readline.GetConfig()
		newConfig.AutoComplete = replCtx.completer
		if err := replCtx.readline.SetConfig(newConfig); err != nil {
			return errors.WrapPrefix(err, "autocomplete reset error", 0)
		}
		replCtx.completer = nil
	}
	return nil
}

func Run(config configuration.Configuration, provider llm.ProviderIfc) error {
	var completer = readline.NewPrefixCompleter(
		// Run commands
		readline.PcItem("/c",
			rlCmdCompleter()...,
		),
		// Help menu
		readline.PcItem("/h"),
		// Quit this program
		readline.PcItem("/q"),
	)
	rl, err := readline.NewFromConfig(&readline.Config{
		AutoComplete:        completer,
		FuncFilterInputRune: filterInput,
		Prompt:              prompts.FirstLine,
	})
	if err != nil {
		return errors.WrapPrefix(err, "failed to create readline", 0)
	}

	modelName := config.String("model")
	model, err := provider.GetModel(context.Background(), modelName)
	if err != nil {
		return errors.WrapPrefix(err, fmt.Sprintf("failed to load model [%s]", modelName), 0)
	}
	replCtx := &ReplContext{
		inputBuffer: new(strings.Builder),
		client:      model,
		session:     llm.Conversation{},
		readline:    rl,
		modelName:   modelName,
	}

	for !replCtx.stopRepl {
		cmd := replCtx.ParseLine()
		if err := cmd.Execute(); err != nil {
			_ = NewPrintErrCmd(replCtx, err).Execute()
		}
	}

	return nil
}

var prompts = struct {
	FirstLine   string
	EmptyPrompt string
}{
	FirstLine:   "[User]: ",
	EmptyPrompt: "",
}

func rlCmdCompleter() []*readline.PrefixCompleter {
	r := make([]*readline.PrefixCompleter, 0)
	r = append(r, readline.PcItem("history"))
	return r
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

type (
	// CmdIfc is the interface of a command object
	CmdIfc interface {
		Execute() error
	}
)
