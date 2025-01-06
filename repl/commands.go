package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"jcheng.org/jcllm/llm"
	"strings"
	"time"
)

// NoOpCmd does nothing, used when the input is an empty line
type NoOpCmd struct {
	*ReplContext
}

func (cmd NoOpCmd) Execute() error {
	return nil
}

// QuitCmd will quit the repl
type QuitCmd struct {
	*ReplContext
}

func NewQuitCmd(ctx *ReplContext) *QuitCmd {
	return &QuitCmd{ReplContext: ctx}
}

func (cmd QuitCmd) Execute() error {
	cmd.stopRepl = true
	return nil
}

// PrintErrCmd prints the error
type PrintErrCmd struct {
	*ReplContext
	err error
}

func (cmd PrintErrCmd) Execute() error {
	fmt.Printf("<Error>%s</Error>\n", strings.TrimRight(cmd.err.Error(), "\n"))
	cmd.ReplContext.ResetInput()
	return nil
}

// AppendCmd appends text the input buffer.
type AppendCmd struct {
	*ReplContext
	line string
}

func NewAppendCmd(replCtx *ReplContext, text string) *AppendCmd {
	return &AppendCmd{replCtx, text}
}

func (cmd AppendCmd) Execute() error {
	cmd.inputBuffer.WriteString(cmd.line)
	cmd.inputBuffer.WriteRune('\n')
	return nil
}

type ChangePromptCmd struct {
	*ReplContext
	prompt string
}

func NewChangePromptCmd(replCtx *ReplContext, prompt string) *ChangePromptCmd {
	return &ChangePromptCmd{replCtx, prompt}
}

func (cmd ChangePromptCmd) Execute() error {
	cmd.ReplContext.readline.SetPrompt(cmd.prompt)
	return nil
}

// SubmitCmd takes the pending input from a multi-line input and submit it to a LLM service for processing.
type SubmitCmd struct {
	*ReplContext
}

func NewSubmitCmd(replCtx *ReplContext) *SubmitCmd {
	return &SubmitCmd{replCtx}
}

func (cmd SubmitCmd) Execute() error {
	stime := time.Now()
	cmd.session.Entries = append(cmd.session.Entries, llm.ChatEntry{
		Role: llm.RoleUser,
		Text: cmd.inputBuffer.String(),
	})

	// Call the LLM REST API
	resp, err := cmd.client.SolicitResponse(context.Background(), llm.Conversation{
		Model:   cmd.modelName,
		Entries: cmd.session.Entries,
	})
	if err != nil {
		return fmt.Errorf("llm client error: %w", err)
	}

	// Prints the "[$model/Assistant]" prompt which precedes the LLM response
	fmt.Printf(fmt.Sprintf("[%s/Assistant]\n", cmd.modelName))
	for elem := range resp.ResponseStream {
		if elem.Err != nil {
			if errors.Is(elem.Err, io.EOF) {
				break
			}
			return fmt.Errorf("response stream error: %w", elem.Err)
		}
		fmt.Print(elem.Text)
	}
	elapsedTime := time.Since(stime)
	fmt.Println("\nElapsed time:", elapsedTime)

	cmd.inputBuffer.Reset()
	cmd.readline.SetPrompt(prompts.FirstLine)

	return nil
}

// ChainCmd chains mulitple commands
type ChainCmd struct {
	cmds []CmdIfc
}

func NewChainCmd(commands ...CmdIfc) *ChainCmd {
	return &ChainCmd{commands}
}

func (cmd ChainCmd) Execute() error {
	for _, cmdIfc := range cmd.cmds {
		if err := cmdIfc.Execute(); err != nil {
			return err
		}
	}
	return nil
}

// EnterMultiLineModeCmd prepares the REPL for multi-line mode.
//  1. The prompt is removed (set to empty string).
//  2. The autocompleter  is stored in the completer member, and the readline autocompleter is turned off.
type EnterMultiLineModeCmd struct {
	*ReplContext
}

func NewEnterMultiLineModeCmd(replCtx *ReplContext) *EnterMultiLineModeCmd {
	return &EnterMultiLineModeCmd{replCtx}
}

func (cmd EnterMultiLineModeCmd) Execute() error {
	readline := cmd.ReplContext.readline
	readline.SetPrompt(prompts.EmptyPrompt)
	configCopy := readline.GetConfig()
	cmd.ReplContext.completer = configCopy.AutoComplete
	configCopy.AutoComplete = nil
	if err := readline.SetConfig(configCopy); err != nil {
		return err
	}
	return nil
}
