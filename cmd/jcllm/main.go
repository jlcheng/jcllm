package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-errors/errors"
	"github.com/jlcheng/jcllm/cli"
	"github.com/jlcheng/jcllm/configuration"
	"github.com/jlcheng/jcllm/configuration/providers/defaultconfig"
)

var (
	version = "undefined"
	commit  = "undefined"
)

func main() {
	config, err := defaultconfig.New(cli.ConfigMetadata, cli.ConfigBools)
	if err != nil {
		if errors.Is(err, configuration.ErrHelp) {
			os.Exit(0)
		}
		fmt.Println(err)
		os.Exit(1)
	}
	app := cli.New(version, commit, config)
	if err := app.Do(); err != nil {
		log.Fatal(err)
	}
}
