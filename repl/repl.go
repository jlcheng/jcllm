package repl

import (
	"context"
	"fmt"
	"github.com/ergochat/readline"
	"github.com/go-errors/errors"
	"io"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/configuration/keys"
	"jcheng.org/jcllm/dye"
	"jcheng.org/jcllm/llm"
	"jcheng.org/jcllm/log"
	"strings"
)

const MultiLinePrefix = "..."

type ReplContext struct {
	config          configuration.Configuration
	logger          *log.Logger
	stopRepl        bool
	inputBuffer     *strings.Builder
	provider        llm.ProviderIfc
	modelName       string
	model           llm.ModelIfc
	session         llm.Conversation
	readline        *readline.Instance
	completer       readline.AutoCompleter
	enableGrounding bool
}

func New(config configuration.Configuration, provider llm.ProviderIfc) (*ReplContext, error) {
	replCtx := &ReplContext{
		stopRepl:    false,
		inputBuffer: new(strings.Builder),
		config:      config,
		provider:    provider,
		logger:      log.New(config.String(keys.OptionLogFile)),
	}

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
	readlineInstance, err := readline.NewFromConfig(&readline.Config{
		AutoComplete:        completer,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return nil, errors.WrapPrefix(err, "failed to create readline", 0)
	}
	replCtx.readline = readlineInstance
	return replCtx, nil
}

func (replCtx *ReplContext) SetModel(modelName string) error {
	model, err := replCtx.provider.GetModel(context.Background(), modelName)
	if err != nil {
		return err
	}
	replCtx.model = model
	replCtx.modelName = modelName
	return nil
}

func (replCtx *ReplContext) ParseLine() CmdIfc {
	slashCommandParser := NewSlashCommandParser(replCtx)
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
		// If this is the first line, try to parse it with the slashCommandParser
		if cmd := slashCommandParser.Parse(line); cmd != nil {
			return cmd
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

	replCtx.prompt(prompts.FirstLine)
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
	replCtx, err := New(config, provider)
	if err != nil {
		return errors.WrapPrefix(err, "failed to create replCtx", 0)
	}
	replCtx.prompt(prompts.FirstLine)
	modelName := config.String(keys.OptionModel)
	if err := replCtx.SetModel(modelName); err != nil {
		return errors.WrapPrefix(err, fmt.Sprintf("failed to set model [%s]", modelName), 0)
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

func (replCtx *ReplContext) prompt(newPrompt string) {
	if newPrompt == prompts.FirstLine {
		newPrompt = dye.Str(newPrompt).Bold().Green()
	}
	replCtx.readline.SetPrompt(newPrompt)
}

func rlCmdCompleter() []*readline.PrefixCompleter {
	r := make([]*readline.PrefixCompleter, 0)
	r = append(r, readline.PcItem("history"))
	r = append(r, readline.PcItem("clear"))
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
