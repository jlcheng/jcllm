package repl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ergochat/readline"
	"github.com/go-errors/errors"
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/keys"
	"github.com/jlcheng/jcllm/dye"
	"github.com/jlcheng/jcllm/llm"
	"github.com/jlcheng/jcllm/log"
)

const MultiLinePrefix = "..."

type ReplContext struct {
	config                  configuration.Configuration
	logger                  *log.Logger
	stopRepl                bool
	inputBuffer             *strings.Builder
	provider                llm.ProviderIfc
	modelName               string
	session                 llm.Conversation
	readline                *readline.Instance
	completer               readline.AutoCompleter
	cmdDefinitions          CommandsProvider
	isMultiLineInputEnabled bool
	solicitResponseArgs     map[string]string
}

func New(config configuration.Configuration, provider llm.ProviderIfc) (*ReplContext, error) {
	replCtx := &ReplContext{
		stopRepl:            false,
		inputBuffer:         new(strings.Builder),
		config:              config,
		provider:            provider,
		logger:              log.New(config.String(keys.OptionLogFile)),
		solicitResponseArgs: make(map[string]string),
	}
	replCtx.cmdDefinitions = newCmdProviderImpl(replCtx)

	var completer = readline.NewPrefixCompleter(
		// Make this a multi-line input
		readline.PcItem("..."),
		// Run commands
		replCtx.slashCommandCompletions(),
		// Help menu
		readline.PcItem("/help"),
		// Quit this program
		readline.PcItem("/quit"),
		replCtx.slashModelCompletions(),
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
	replCtx.modelName = modelName
	replCtx.UpdatePrompt()
	return nil
}

func (replCtx *ReplContext) ParseLine() CmdIfc {
	slashCommandParser := NewSlashCommandParser(replCtx)
	line, err := replCtx.readline.Readline()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return NewQuitCmd(replCtx)
		} else {
			return NewPrintErrCmd(replCtx, err)
		}
	}
	if len(line) == 0 {
		return NewNoOpCmd()
	}

	if replCtx.inputBuffer.Len() == 0 {
		if strings.TrimSpace(line) == "/quit" {
			return NewQuitCmd(replCtx)
		}

		if strings.TrimSpace(line) == "/help" {
			return NewHelpCmd(replCtx)
		}

		// If this is the first line, try to parse it with the slashCommandParser
		if cmd := slashCommandParser.Parse(line); cmd != nil {
			return cmd
		}

		// Handle the /m command to change the model
		if strings.HasPrefix(line, "/m ") {
			modelName := strings.TrimSpace(strings.TrimPrefix(line, "/m "))
			return NewSetModelCmd(replCtx, modelName)
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

	replCtx.SetMultiLineInput(false)
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

func (replCtx *ReplContext) Close() {
	if replCtx.readline != nil {
		if err := replCtx.readline.Close(); err != nil {
			replCtx.logger.Errorf("failed to close readline: %v", err)
			fmt.Printf("failed to close readline: %v\n", err)
		}
	}
}

func Run(config configuration.Configuration, provider llm.ProviderIfc) error {
	replCtx, err := New(config, provider)
	if err != nil {
		return errors.WrapPrefix(err, "failed to create replCtx", 0)
	}
	defer replCtx.Close()
	modelName := config.String(keys.OptionModel)
	if err := replCtx.SetModel(modelName); err != nil {
		return errors.WrapPrefix(err, fmt.Sprintf("failed to set model [%s]", modelName), 0)
	}
	replCtx.SetMultiLineInput(false)

	for !replCtx.stopRepl {
		cmd := replCtx.ParseLine()
		if err := cmd.Execute(); err != nil {
			_ = NewPrintErrCmd(replCtx, err).Execute()
		}
	}

	return nil
}

func (replCtx *ReplContext) UpdatePrompt() {
	if replCtx.isMultiLineInputEnabled {
		replCtx.readline.SetPrompt("")
		return
	}
	promptPrefix := dye.Str("[To ").Green()
	modelName := dye.Str(replCtx.modelName).Bold().Yellow()
	promptSuffix := dye.Str("]:").Green()
	formattedPrompt := fmt.Sprintf("%s%s%s ", promptPrefix, modelName, promptSuffix)
	replCtx.readline.SetPrompt(formattedPrompt)
}

func (replCtx *ReplContext) SetMultiLineInput(isMultiLineInputEnabled bool) {
	replCtx.isMultiLineInputEnabled = isMultiLineInputEnabled
	replCtx.UpdatePrompt()
}

func (replCtx *ReplContext) slashCommandCompletions() *readline.PrefixCompleter {
	cmdMap := newCmdProviderImpl(replCtx).Commands()
	r := make([]*readline.PrefixCompleter, 0)
	for cmdName := range cmdMap {
		r = append(r, readline.PcItem(cmdName))
	}
	return readline.PcItem("/c", r...)
}

func (replCtx *ReplContext) slashModelCompletions() *readline.PrefixCompleter {
	providerName := replCtx.config.String(keys.OptionProvider)
	modelsListKey := fmt.Sprintf("%s-%s", providerName, keys.OptionModelsList)
	models := replCtx.config.Strings(modelsListKey)
	if len(models) == 0 {
		fetchedModels, err := replCtx.provider.ListModels(context.Background())
		if err != nil {
			models = []string{"<error>cannot fetch models</error>"}
		} else {
			models = make([]string, len(fetchedModels))
			for idx, model := range fetchedModels {
				models[idx] = model.Name
			}
		}
	}

	return readline.PcItemDynamic(func(_ string) []string {
		var items []string
		for _, model := range models {
			items = append(items, fmt.Sprintf("/m %s", model))
		}
		return items
	})
}

type CommandsProvider interface {
	Commands() map[string]CmdIfc
}

type cmdProviderImpl struct {
	replCtx *ReplContext
}

func newCmdProviderImpl(replCtx *ReplContext) *cmdProviderImpl {
	return &cmdProviderImpl{replCtx: replCtx}
}

func (impl *cmdProviderImpl) Commands() map[string]CmdIfc {
	return map[string]CmdIfc{
		"history":  NewSummarizeHistoryCmd(impl.replCtx),
		"clear":    NewClearConversationCommand(impl.replCtx),
		"suppress": NewSuppressCommand(impl.replCtx),
	}
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
