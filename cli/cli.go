package cli

import (
	"context"
	"fmt"

	"github.com/go-errors/errors"
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/keys"
	"github.com/jlcheng/jcllm/llm"
	"github.com/jlcheng/jcllm/llm/providers/registry"
	"github.com/jlcheng/jcllm/log"
	"github.com/jlcheng/jcllm/repl"
)

type CLI struct {
	version string
	commit  string
	config  configuration.Configuration
	logger  *log.Logger
}

func New(version string, commit string, config configuration.Configuration) *CLI {
	return &CLI{
		version: version,
		commit:  commit,
		config:  config,
		logger:  log.New(config.String(keys.OptionLogFile)),
	}
}

func (cli *CLI) ListProviders() error {
	providers := []string{keys.ProviderGeminiLegacy}
	fmt.Println("Supported providers:")
	for _, provider := range providers {
		fmt.Println(provider)
	}
	return nil
}

func (cli *CLI) ListModels() error {
	name := cli.config.String(keys.OptionProvider)
	provider, err := registry.NewProvider(context.Background(), cli.config, name)
	if err != nil {
		cli.logger.Errorf("cannot instantiate provider [%s]: %v", name, err)
		if errors.Is(err, llm.ErrProviderNotFound) {
			fmt.Printf("provider not found: %v\n", name)
			return nil
		}
		return errors.WrapPrefix(err, fmt.Sprintf("cannot instantiate provider [%s]", name), 0)
	}
	models, err := provider.ListModels(context.Background())
	if err != nil {
		cli.logger.Errorf("cannot list models: %v", err)
		return errors.Errorf("cannot list models: %v", err)
	}
	for _, model := range models {
		fmt.Printf("=== %s ===\n", model.Name)
		fmt.Printf("    Description: %s\n", model.Description)
		fmt.Printf("    Max tokens: %d\n", model.MaxTokens)
		fmt.Printf("    Version: %s\n", model.Version)
	}
	return nil
}

func (cli *CLI) Repl() error {
	fmt.Printf("jcllm version: %s\n", cli.version)
	name := cli.config.String(keys.OptionProvider)
	provider, err := registry.NewProvider(context.Background(), cli.config, name)
	if err != nil {
		return errors.WrapPrefix(err, "provider error", 0)
	}

	if err := repl.Run(cli.config, provider); err != nil {
		return errors.WrapPrefix(err, "repl error", 0)
	}
	return nil
}

func (cli *CLI) Do() error {
	if cli.config.Bool("version") {
		fmt.Printf("jcllm version: %s\n", cli.version)
		fmt.Printf("commit: %s\n", cli.commit)
		return nil
	}

	command := cli.config.String(keys.OptionCommand)
	switch command {
	case "list-models":
		if err := cli.ListModels(); err != nil {
			cli.logger.Errorf("cannot list models: %v", err)
			return err
		}
	case "list-providers":
		if err := cli.ListProviders(); err != nil {
			cli.logger.Errorf("cannot list providers: %v", err)
			return err
		}
	case "repl":
		if err := cli.Repl(); err != nil {
			cli.logger.Errorf("cannot start repl: %v", err)
			return err
		}
	case "":
		return fmt.Errorf("no command specified")
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}
