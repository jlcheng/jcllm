package cli

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/llm"
	"jcheng.org/jcllm/llm/providers/registry"
	"log"
)

type CLI struct {
	config configuration.Configuration
}

func New(config configuration.Configuration) *CLI {
	return &CLI{config: config}
}

func (cli *CLI) ListProviders() error {
	providers := []string{"google"}
	fmt.Println("Supported providers:")
	for _, provider := range providers {
		fmt.Println(provider)
	}
	return nil
}

func (cli *CLI) ListModels() error {
	name := cli.config.String("provider")
	provider, err := registry.NewProvider(context.Background(), cli.config, name)
	if err != nil {
		if errors.Is(err, llm.ErrNotFound) {
			log.Fatalf("provider not found: %v", name)
		}
		log.Fatalf("cannot instantiate provider [%s]: %v ", name, err)
	}
	models, err := provider.ListModels(context.Background())
	if err != nil {
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

func (cli *CLI) Do() error {
	switch cli.config.String("command") {
	case "list-models":
		if err := cli.ListModels(); err != nil {
			return err
		}
	case "list-providers":
		if err := cli.ListProviders(); err != nil {
			return err
		}
	case "":
		return fmt.Errorf("command not specified")
	default:
		return fmt.Errorf("unknown command: %s", cli.config.String("command"))
	}
	return nil
}
