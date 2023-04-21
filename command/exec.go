package command

import (
	"fmt"
	"os"

	"github.com/golang108/drone-runner-exec/command/internal"
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
	fmt.Println("executes a pipeline func run.....")
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
