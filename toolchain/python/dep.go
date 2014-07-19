package python

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"sourcegraph.com/sourcegraph/srcgraph/config"
	"sourcegraph.com/sourcegraph/srcgraph/container"
	"sourcegraph.com/sourcegraph/srcgraph/dep2"
	"sourcegraph.com/sourcegraph/srcgraph/task2"
	"sourcegraph.com/sourcegraph/srcgraph/unit"
)

func init() {
	dep2.RegisterLister(&FauxPackage{}, dep2.DockerLister{defaultPythonEnv})
	dep2.RegisterResolver(pythonRequirementTargetType, defaultPythonEnv)
}

func (p *pythonEnv) BuildLister(dir string, unit unit.SourceUnit, c *config.Repository, x *task2.Context) (*container.Command, error) {
	dockerfile, err := p.depDockerfile()
	if err != nil {
		return nil, err
	}

	return &container.Command{
		Container: container.Container{
			Dockerfile: dockerfile,
			RunOptions: []string{"-v", dir + ":" + srcRoot},
			Cmd:        []string{"pydep-run.py", srcRoot},
			Stderr:     x.Stderr,
			Stdout:     x.Stdout,
		},
		Transform: func(orig []byte) ([]byte, error) {
			var reqs []requirement
			err := json.NewDecoder(bytes.NewReader(orig)).Decode(&reqs)
			if err != nil {
				return nil, err
			}
			deps := make([]*dep2.RawDependency, len(reqs))
			for i, req := range reqs {
				deps[i] = &dep2.RawDependency{
					TargetType: pythonRequirementTargetType,
					Target:     req,
				}
			}
			return json.Marshal(deps)
		},
	}, nil
}

type requirement struct {
	ProjectName string      `json:"project_name"`
	UnsafeName  string      `json:"unsafe_name"`
	Key         string      `json:"key"`
	Specs       [][2]string `json:"specs"`
	Extras      []string    `json:"extras"`
	RepoURL     string      `json:"repo_url"`
	Packages    []string    `json:"packages"`
	Modules     []string    `json:"modules"`
}

func (p *pythonEnv) Resolve(dep *dep2.RawDependency, c *config.Repository, x *task2.Context) (*dep2.ResolvedTarget, error) {
	switch dep.TargetType {
	case pythonRequirementTargetType:
		var req requirement
		reqJson, _ := json.Marshal(dep.Target)
		json.Unmarshal(reqJson, &req)

		toUnit := &FauxPackage{}
		return &dep2.ResolvedTarget{
			ToRepoCloneURL: req.RepoURL,
			ToUnit:         toUnit.Name(),
			ToUnitType:     unit.Type(toUnit),
		}, nil
	default:
		return nil, fmt.Errorf("Unexpected target type for Python: %+v", dep.TargetType)
	}
}

func (l *pythonEnv) depDockerfile() ([]byte, error) {
	var buf bytes.Buffer
	template.Must(template.New("").Parse(depDockerfile)).Execute(&buf, l)
	return buf.Bytes(), nil
}

const pythonRequirementTargetType = "python-requirement"
const depDockerfile = `FROM ubuntu:14.04
RUN apt-get update
RUN apt-get install -qy curl
RUN apt-get install -qy git
RUN apt-get install -qy {{.PythonVersion}}
RUN ln -s $(which {{.PythonVersion}}) /usr/bin/python
RUN apt-get install -qy python-pip

RUN pip install git+git://github.com/sourcegraph/pydep@{{.PydepVersion}}
`
