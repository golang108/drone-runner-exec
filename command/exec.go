package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/golang108/drone-runner-exec/command/internal"
	"github.com/golang108/drone-runner-exec/engine"
	"github.com/golang108/drone-runner-exec/engine/compiler"
	"github.com/golang108/drone-runner-exec/engine/resource"
	"github.com/golang108/drone-runner-exec/runtime"

	"github.com/drone/drone-go/drone"
	"github.com/drone/envsubst"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline"
	"github.com/drone/runner-go/pipeline/console"
	"github.com/drone/runner-go/secret"
	"github.com/drone/signal"

	"github.com/mattn/go-isatty"
	"gopkg.in/alecthomas/kingpin.v2"
)

type execCommand struct {
	*internal.Flags

	Root    string
	Source  *os.File
	Environ map[string]string
	Secrets map[string]string
	Pretty  bool
	Procs   int64
}

func (c *execCommand) run(*kingpin.ParseContext) error {
	rawsource, err := io.ReadAll(c.Source)
	if err != nil {
		return err
	}

	envs := environ.Combine(
	// empty envs
	)
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

	// parse and lint the configuration.
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
		Pipeline: resource, // 这个是通过yml中name获取的那个
		Manifest: manifest, // 这里面保存了全部的resource
		Build:    c.Build,
		Netrc:    nil, // c.Netrc,  // 新建 netrc 文件。这个莫名其妙的 这里直接干掉
		Repo:     c.Repo,
		Stage:    c.Stage,
		System:   c.System,
		Environ:  c.Environ,
		Secret:   secret.StaticVars(c.Secrets),
		Root:     c.Root, // 这个来自命令行
	}
	spec := comp.Compile(nocontext)

	fmt.Println(spec)
	// create a step object for each pipeline step.
	for _, step := range spec.Steps {
		if step.RunPolicy == engine.RunNever {
			continue
		}
		c.Stage.Steps = append(c.Stage.Steps, &drone.Step{
			StageID:   c.Stage.ID,
			Number:    len(c.Stage.Steps) + 1,
			Name:      step.Name,
			Status:    drone.StatusPending,
			ErrIgnore: step.IgnoreErr,
		})
	}

	// configures the pipeline timeout.
	timeout := time.Duration(c.Repo.Timeout) * time.Minute
	ctx, cancel := context.WithTimeout(nocontext, timeout)
	defer cancel()

	// listen for operating system signals and cancel execution
	// when received.
	ctx = signal.WithContextFunc(ctx, func() {
		println("received signal, terminating process")
		cancel()
	})

	state := &pipeline.State{
		Build:  c.Build,
		Stage:  c.Stage,
		Repo:   c.Repo,
		System: c.System,
	}
	err = runtime.NewExecer(
		pipeline.NopReporter(),
		console.New(c.Pretty),
		engine.New(),
		c.Procs,
	).Exec(ctx, spec, state)
	if err != nil {
		return err
	}
	switch state.Stage.Status {
	case drone.StatusError, drone.StatusFailing:
		os.Exit(1)
	}

	return nil
}

func registerExec(app *kingpin.Application) {
	c := new(execCommand)

	cmd := app.Command("exec", "executes a pipeline").
		Action(c.run)

	cmd.Arg("root", "root build directory").
		Default("").
		StringVar(&c.Root)

	cmd.Arg("source", "source file location").
		Default(".drone.yml").
		FileVar(&c.Source)

	cmd.Flag("pretty", "pretty print the output").
		Default(
			fmt.Sprint(
				isatty.IsTerminal(
					os.Stdout.Fd(),
				),
			),
		).BoolVar(&c.Pretty)

	// shared pipeline flags
	c.Flags = internal.ParseFlags(cmd)
}
