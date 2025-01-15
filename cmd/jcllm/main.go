package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"jcheng.org/jcllm/cli"
	"jcheng.org/jcllm/configuration"
	"jcheng.org/jcllm/configuration/providers/defaultconfig"
	"log"
	"os"
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
