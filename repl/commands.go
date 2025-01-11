package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"jcheng.org/jcllm/dye"
	"jcheng.org/jcllm/llm"
	"math"
	"strings"
	"time"
)

type (
	// CmdIfc is the interface of a command object
	CmdIfc interface {
		Execute() error
	}
)

// NewNoOpCmd creates a command which does nothing, useful when a user mistakenly enters a blank line.
func NewNoOpCmd() CmdIfc {
	return NewLambdaCmd(func() error {
		return nil
	})
}

// NewQuitCmd creates a command to quit the REPL.
func NewQuitCmd(replCtx *ReplContext) CmdIfc {
	return NewLambdaCmd(func() error {
		replCtx.stopRepl = true
		return nil
	})
}

// NewPrintErrCmd creates a command which prints the given error.
func NewPrintErrCmd(replCtx *ReplContext, err error) CmdIfc {
	return NewLambdaCmd(func() error {
		if err := replCtx.ResetInput(); err != nil {
			fmt.Printf("<Error>An error occured when resetting the input buffer: %s</Error>\n", err.Error())
		}
		fmt.Printf("<Error>%s</Error>\n", strings.TrimRight(err.Error(), "\n"))
		return nil
	})
}

// NewAppendCmd creates a command which appends the text to the input buffer.
func NewAppendCmd(replCtx *ReplContext, text string) CmdIfc {
	return NewLambdaCmd(func() error {
		replCtx.inputBuffer.WriteString(text)
		replCtx.inputBuffer.WriteRune('\n')
		return nil
	})
}

// NewSubmitCmd creates a command which takes the pending input and submit it to a LLM for processing.
func NewSubmitCmd(replCtx *ReplContext) CmdIfc {
	return NewLambdaCmd(func() error {
		startTime := time.Now()
		session := &replCtx.session
		session.Entries = append(session.Entries, llm.ChatEntry{
			Role: llm.RoleUser,
			Text: replCtx.inputBuffer.String(),
		})

		resp, err := replCtx.model.SolicitResponse(context.Background(), llm.Conversation{
			Entries: session.Entries,
		})
		if err != nil {
			return fmt.Errorf("llm client error: %w", err)
		}
		var responseBuffer strings.Builder

		tokens := 0
		fmt.Println(dye.Strf("[%s]:", replCtx.modelName).Bold().Yellow())
		for elem := range resp.ResponseStream {
			if elem.Err != nil {
				if errors.Is(elem.Err, io.EOF) {
					break
				}
				return fmt.Errorf("response stream error: %w", elem.Err)
			}
			// Print out each token as soon as it arrives
			fmt.Print(elem.Text)
			responseBuffer.WriteString(elem.Text)
			tokens += elem.TokenCount
		}
		elapsedTime := time.Since(startTime)
		tokensPerSec := float64(tokens) / math.Max(1, elapsedTime.Seconds())
		fmt.Printf("\n[%.2f tokens/s, %.2fs, %d tokens]\n", tokensPerSec, elapsedTime.Seconds(), tokens)
		session.Entries = append(session.Entries, llm.ChatEntry{
			Role: llm.RoleAssistant,
			Text: responseBuffer.String(),
		})
		if err := replCtx.ResetInput(); err != nil {
			return fmt.Errorf("input reset error: %w", err)
		}
		return nil
	})
}

// NewChainCmd creates a command which runs multiple commands in sequence.
func NewChainCmd(commands ...CmdIfc) CmdIfc {
	return NewLambdaCmd(func() error {
		for _, cmdIfc := range commands {
			if err := cmdIfc.Execute(); err != nil {
				return err
			}
		}
		return nil
	})
}

// NewEnterMultiLineModeCmd creates a command which prepares the REPL for multi-line mode.
//  1. The prompt is removed (set to empty string).
//  2. Auto-complete is turned off. A copy of the auto-complete function is stored so it can be re-enabled later.
func NewEnterMultiLineModeCmd(replCtx *ReplContext) CmdIfc {
	return NewLambdaCmd(func() error {
		readline := replCtx.readline
		replCtx.prompt(prompts.EmptyPrompt)
		configCopy := readline.GetConfig()
		replCtx.completer = configCopy.AutoComplete
		configCopy.AutoComplete = nil
		if err := readline.SetConfig(configCopy); err != nil {
			return err
		}
		return nil
	})
}

type LambdaCmd struct {
	executeFunction func() error
}

func NewLambdaCmd(executeFunction func() error) *LambdaCmd {
	return &LambdaCmd{executeFunction}
}

func (cmd *LambdaCmd) Execute() error {
	return cmd.executeFunction()
}

// NewSummarizeHistoryCmd creates a command which prints the current conversation.
func NewSummarizeHistoryCmd(replCtx *ReplContext) CmdIfc {
	roleToPrefix := func(entry llm.ChatEntry) string {
		switch entry.Role {
		case llm.RoleUser:
			return fmt.Sprintf("%15s", "[User]: ")
		case llm.RoleAssistant:
			return fmt.Sprintf("%15s", "[Assistant]: ")
		default:
			return "[Unknown]: "
		}
	}
	summarizeText := func(entry llm.ChatEntry) string {
		summarized := strings.TrimSpace(entry.Text)
		summarized = strings.ReplaceAll(summarized, "\n", "¶ ")
		maxLength := 80
		suffix := "..."
		if len(summarized) > maxLength {
			return summarized[:maxLength-len(suffix)] + suffix
		}
		return summarized
	}

	return NewLambdaCmd(func() error {
		chatEntries := replCtx.session.Entries
		fmt.Println("=== Conversation Summary ===")
		for _, chatEntry := range chatEntries {
			fmt.Println(roleToPrefix(chatEntry) + " " + summarizeText(chatEntry))
		}
		fmt.Println("======= End Summary ========")
		return nil
	})
}

func NewClearConversationCommand(replCtx *ReplContext) CmdIfc {
	return NewLambdaCmd(func() error {
		replCtx.session.Entries = replCtx.session.Entries[:0]
		fmt.Println(dye.Str("[Current conversation cleared]").Bold().Yellow())
		if err := replCtx.ResetInput(); err != nil {
			return err
		}
		return nil
	})
}
