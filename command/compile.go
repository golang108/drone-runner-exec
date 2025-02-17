package command

import (
	"encoding/json"
	"fmt"
	"github.com/drone/envsubst"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/secret"
	"github.com/golang108/drone-runner-exec/command/internal"
	"github.com/golang108/drone-runner-exec/engine/compiler"
	"github.com/golang108/drone-runner-exec/engine/resource"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"strings"
)

type compileCommand struct {
	*internal.Flags

	Root    string
	Source  *os.File
	Environ map[string]string
	Secrets map[string]string
}

func (c *compileCommand) run(*kingpin.ParseContext) error {
	rawsource, err := io.ReadAll(c.Source)
	if err != nil {
		return err
	}

	envs := environ.Combine()

	// string substitution function ensures that string
	// replacement variables are escaped and quoted if they
	// contain newlines.
	subf := func(k string) string {
		v := envs[k]
		if strings.Contains(v, "\n") {
			v = fmt.Sprintf("%q", v)
		}
		return v
	}

	// evaluates string replacement expressions and returns an
	// update configuration.
	config, err := envsubst.Eval(string(rawsource), subf)
	if err != nil {
		return err
	}

	// parse and lint the configuration
	manifest, err := manifest.ParseString(config)
	if err != nil {
		return err
	}

	// a configuration can contain multiple pipelines.
	// get a specific pipeline resource for execution.
	resource, err := resource.Lookup(c.Stage.Name, manifest)
	if err != nil {
		return err
	}

	// compile the pipeline to an intermediate representation.
	comp := &compiler.Compiler{
		Pipeline: resource,
		Manifest: manifest,
		Build:    c.Build,
		Netrc:    c.Netrc,
		Repo:     c.Repo,
		Stage:    c.Stage,
		System:   c.System,
		Environ:  c.Environ,
		Secret:   secret.StaticVars(c.Secrets),
		Root:     c.Root,
	}
	spec := comp.Compile(nocontext)

	// encode the pipeline in json format and print to the
	// console for inspection.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(spec)
	return nil
}

func registerCompile(app *kingpin.Application) {
	c := new(compileCommand)

	cmd := app.Command("compile", "compile the yaml file").
		Action(c.run)

	cmd.Arg("root", "root build directory").
		Default("").
		StringVar(&c.Root)

	cmd.Arg("source", "source file location").
		Default(".drone.yml").
		FileVar(&c.Source)

	// shared pipeline flags
	c.Flags = internal.ParseFlags(cmd)
}
