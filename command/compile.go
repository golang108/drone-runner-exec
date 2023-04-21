package command

import (
	"fmt"
	"github.com/golang108/drone-runner-exec/command/internal"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

type compileCommand struct {
	*internal.Flags

	Root    string
	Source  *os.File
	Environ map[string]string
	Secrets map[string]string
}

func (c *compileCommand) run(*kingpin.ParseContext) error {
	fmt.Println("compile the yaml file func run .....")
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
