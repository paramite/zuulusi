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
		JobID: "3f8e570ca9144a79b30539ab021388f9/controller/ci-framework-data/logs/openstack-k8s-operators-openstack-must-gather/namespaces",
	})
	if err != nil {
		log.Fatal("could not generate model response: ", err)
	}

	fmt.Printf(resp)
}
