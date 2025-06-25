package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/paramite/zuulusi/agents"
	"github.com/paramite/zuulusi/tools"
)

var k = koanf.New(".")

func main() {
	// read config location from env or cmdline
	cfg := os.Getenv("ZUULUSI_CONFIG")
	if cfg == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("failed to get current user: %v", err)
		}
		hdir := usr.HomeDir
		flag.StringVar(&cfg, "config", fmt.Sprintf("%s/.config/zuulusi/config.yaml", hdir),
			"Path to a configuration file")
		flag.Parse()
	}

	if err := k.Load(file.Provider(cfg), yaml.Parser()); err != nil {
		log.Fatalf("failed to load configuration file: %v", err)
	}

	if len(k.Strings("prompt.important")) < 1 {
		log.Fatalf("Missing config value of prompt.important")
	}

	if len(k.String("prompt.ack_with_occurences")) < 1 {
		log.Fatalf("Missing config value of prompt.ack_with_occurences")
	}

	if len(os.Args) < 2 {
		log.Fatalf("At least one argument with value of Job ID is required.\n")
	}
	jobID := os.Args[1]

	ctx := context.Background()
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: k.String("google.api_key")}),
		genkit.WithDefaultModel("googleai/gemini-2.0-flash"),
	)
	if err != nil {
		log.Fatalf("failed to initialize Genkit: %v", err)
	}

	genkit.DefineTool(g, "mycurl", "Downloads content of a given file.", tools.CurlTool())
	genkit.DefineTool(g, "dircrawler", "Builds list of log files of given CI job.", tools.DirCrawlerTool())

	agent := agents.LogCrawlerAgent(g)

	resp, err := agent.Run(ctx, agents.LogCrawlerInput{
		JobID:             jobID,
		ImportantDirs:     k.Strings("prompt.important"),
		MinimumOccurences: k.String("prompt.ack_with_occurences"),
	})
	if err != nil {
		log.Fatal("could not generate model response: ", err)
	}

	fmt.Printf("\nFound following potential issue(s):\n%s", resp)
}
