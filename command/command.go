package command

import (
	"context"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

// program version
var version = "0.0.0"

// empty context
var nocontext = context.Background()

// Command parses the command line arguments and then executes a
// subcommand program.
func Command() {
	app := kingpin.New("drone", "drone exec runner")
	registerCompile(app)
	registerExec(app)
	//registerDaemon(app)
	//service.Register(app)

	kingpin.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
