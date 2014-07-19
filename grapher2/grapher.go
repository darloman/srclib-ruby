package grapher2

import (
	"encoding/json"
	"fmt"

	"sourcegraph.com/sourcegraph/srcgraph/config"
	"sourcegraph.com/sourcegraph/srcgraph/container"
	"sourcegraph.com/sourcegraph/srcgraph/graph"
	"sourcegraph.com/sourcegraph/srcgraph/task2"
	"sourcegraph.com/sourcegraph/srcgraph/unit"
)

type Grapher interface {
	Graph(dir string, unit unit.SourceUnit, c *config.Repository, x *task2.Context) (*Output, error)
}

type Output struct {
	Symbols []*graph.Symbol `json:",omitempty"`
	Refs    []*graph.Ref    `json:",omitempty"`
	Docs    []*graph.Doc    `json:",omitempty"`
}

type GrapherBuilder interface {
	BuildGrapher(dir string, unit unit.SourceUnit, c *config.Repository, x *task2.Context) (*container.Command, error)
}

type DockerGrapher struct {
	GrapherBuilder
}

func (g DockerGrapher) Graph(dir string, unit unit.SourceUnit, c *config.Repository, x *task2.Context) (*Output, error) {
	cmd, err := g.BuildGrapher(dir, unit, c, x)
	if err != nil {
		return nil, err
	}

	data, err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var output *Output
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}

	// Basic uniqueness checks.
	seenSymbolPaths := make(map[graph.SymbolPath]*graph.Symbol, len(output.Symbols))
	for _, s := range output.Symbols {
		if s0, seen := seenSymbolPaths[s.Path]; seen {
			return nil, fmt.Errorf("duplicate path in symbols output: %q\nsymbol 1: %+v\nsymbol 2: %+v", s.Path, s0, s)
		}
		seenSymbolPaths[s.Path] = s
	}

	return output, nil
}

// Graph uses the registered grapher (if any) to graph the source unit (whose repository is cloned to
// dir).
func Graph(dir string, unit unit.SourceUnit, c *config.Repository, x *task2.Context) (*Output, error) {
	g, registered := Graphers[ptrTo(unit)]
	if !registered {
		return nil, fmt.Errorf("no grapher registered for source unit %T", unit)
	}

	return g.Graph(dir, unit, c, x)
}
