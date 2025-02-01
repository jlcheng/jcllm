package repl

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/go-errors/errors"

	"github.com/jlcheng/jcllm/configuration/keys"
	"github.com/jlcheng/jcllm/dye"
	"github.com/jlcheng/jcllm/llm"
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

func NewHelpCmd(_ *ReplContext) CmdIfc {
	return NewLambdaCmd(func() error {
		fmt.Printf("Special commands:\n")
		fmt.Printf("  %-20sShow this help text\n", "/help")
		fmt.Printf("  %-20sQuits the program\n", "/quit")
		fmt.Printf("  %-20sPrints a summary of the chat history\n", "/c history")
		fmt.Printf("  %-20sClears the chat history\n", "/c clear ")
		fmt.Printf("  %-20sSuppresses the @ground feature when using Gemini\n", "/c suppress")
		fmt.Printf("  %-20sChange models\n", "/m <model_name>")
		fmt.Printf("  %-20sStart with 3 periods (...) to enter multi-line text; End with a single period on its own line\n", "...")
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
		// We want to ensure the suppress command is only applied for one turn of conversation.
		defer func() { replCtx.solicitResponseArgs[keys.ArgNameSuppress] = keys.False }()
		session := &replCtx.session
		session.Entries = append(session.Entries, llm.ChatEntry{
			Role: llm.RoleUser,
			Text: replCtx.inputBuffer.String(),
		})
		// Reset the input states as soon as possible, since there are multiple places where this method might return early
		if err := replCtx.ResetInput(); err != nil {
			return errors.WrapPrefix(err, "input reset failed", 0)
		}

		resp, err := replCtx.provider.SolicitResponse(context.Background(), llm.SolicitResponseInput{
			ModelName: replCtx.modelName,
			Conversation: llm.Conversation{
				Entries: session.Entries,
			},
			Args: replCtx.solicitResponseArgs,
		})
		if err != nil {
			// We allow users to append mentions at the end of the input, e.g., "What happened today. @ground". This means an input with
			// only mentions appear blank _after_ preprocessing. Thus, we need to handle blank inputs again here.
			// Maybe using mentions as a UI element is not a good idea?
			if errors.Is(err, llm.ErrBlankInput) {
				session.Entries = session.Entries[:len(session.Entries)-1]
				return nil
			}
			return errors.WrapPrefix(err, "request to llm failed", 0)
		}
		var responseBuffer strings.Builder

		tokens := 0
		fmt.Println(dye.Strf("[%s]:", replCtx.modelName).Bold().Yellow())
		for message, err := range resp.Messages {
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return errors.WrapPrefix(err, "error read from llm stream", 0)
			}
			// Print out each token as soon as it arrives
			fmt.Print(message.Text)
			responseBuffer.WriteString(message.Text)
			tokens += message.TokenCount
		}
		fmt.Println()
		elapsedTime := time.Since(startTime)
		tokensPerSec := float64(tokens) / math.Max(1, elapsedTime.Seconds())
		fmt.Printf("[%.2f tokens/s, %.2fs, %d tokens]\n", tokensPerSec, elapsedTime.Seconds(), tokens)
		session.Entries = append(session.Entries, llm.ChatEntry{
			Role: llm.RoleAssistant,
			Text: responseBuffer.String(),
		})
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
		replCtx.SetMultiLineInput(true)
		// TODO consider moving this to SetMultiLineInput
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

func NewSetModelCmd(replCtx *ReplContext, modelName string) CmdIfc {
	return NewLambdaCmd(func() error {
		if err := replCtx.SetModel(modelName); err != nil {
			return err
		}
		fmt.Println(dye.Strf("Model set to: %s", modelName).Bold().Yellow())
		return nil
	})
}

func NewSuppressCommand(replCtx *ReplContext) CmdIfc {
	return NewLambdaCmd(func() error {
		replCtx.solicitResponseArgs[keys.ArgNameSuppress] = keys.True
		fmt.Printf("%q\n", replCtx.solicitResponseArgs)
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
