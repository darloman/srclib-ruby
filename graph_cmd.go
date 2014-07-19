package srcgraph

import (
	"flag"
	"fmt"
	"log"
	"os"

	"sourcegraph.com/sourcegraph/srcgraph/grapher2"
	"sourcegraph.com/sourcegraph/srcgraph/task2"
	"sourcegraph.com/sourcegraph/srcgraph/unit"
)

func graph_(args []string) {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "show JSON output")
	summary := fs.Bool("summary", true, "summarize output data")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: `+Name+` graph [options] [unit...]

Analyze a repository's source code for definitions and references. If unit(s)
are specified, only source units with matching IDs will be graphed.

The options are:
`)
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)
	sourceUnitSpecs := fs.Args()

	context, err := NewJobContext(*Dir, task2.DefaultContext)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range context.Repo.SourceUnits {
		if !SourceUnitMatchesArgs(sourceUnitSpecs, u) {
			continue
		}

		output, err := grapher2.Graph(context.RepoRootDir, u, context.Repo, task2.DefaultContext)
		if err != nil {
			log.Fatal(err)
		}

		if *summary || *Verbose {
			log.Printf("## %s output summary:", unit.MakeID(u))
			log.Printf(" - %d symbols", len(output.Symbols))
			log.Printf(" - %d refs", len(output.Refs))
			log.Printf(" - %d docs", len(output.Docs))
		}

		if *jsonOutput {
			PrintJSON(output, "")
		}

		fmt.Println()
	}
}
