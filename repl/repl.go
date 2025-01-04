package repl

import (
	"context"
	"fmt"
	"github.com/ergochat/readline"
	"github.com/go-errors/errors"
	"io"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
	"os"
	"strings"
)

func Run(config configuration.Configuration, provider llm.ProviderIfc) error {
	var completer = readline.NewPrefixCompleter(
		// Run commands
		readline.PcItem("\\c",
			rlCmdCompleter()...,
		),
		// Help menu
		readline.PcItem("\\h"),
		// Quit this program
		readline.PcItem("\\q"),
	)
	rl, err := readline.NewEx(&readline.Config{
		AutoComplete: completer,
		Listener:     keyListener,
	})
	if err != nil {
		return errors.WrapPrefix(err, "failed to create readline", 0)
	}
	rl.SetPrompt("> ")

	modelName := config.String("model")
	model, err := provider.GetModel(context.Background(), modelName)
	if err != nil {
		return errors.WrapPrefix(err, fmt.Sprintf("failed to load model [%s]", modelName), 0)
	}
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.WrapPrefix(err, "failed to read line", 0)
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		respStream, err := model.SolicitResponse(context.Background(), llm.Conversation{
			Model:   modelName,
			Entries: []llm.ChatEntry{{Text: line}},
		})
		if err != nil {
			return errors.WrapPrefix(err, "failed to read response", 0)
		}
		for chunk := range respStream.ResponseStream {
			if chunk.Err != nil {
				return errors.WrapPrefix(chunk.Err, "failed to read response", 0)
			}
			fmt.Print(chunk.Text)
		}
	}
	return nil
}

func rlCmdCompleter() []*readline.PrefixCompleter {
	r := make([]*readline.PrefixCompleter, 0)
	r = append(r, readline.PcItem("option1"))
	return r
}

var f io.Writer = getWriter()

func getWriter() io.Writer {
	f, err := os.OpenFile("/home/jcheng/tmp/tmp.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	return f
}

func filterInput(r rune) (rune, bool) {
	fmt.Fprintf(f, "Got rune [%v]\n", r)
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func keyListener(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	fmt.Fprintf(f, "got key [%v]\n", key)
	return line, pos, true
}
