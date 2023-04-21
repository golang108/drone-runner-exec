package internal

import (
	"fmt"
	"time"

	"github.com/drone/drone-go/drone"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags maps
type Flags struct {
	Build  *drone.Build
	Netrc  *drone.Netrc
	Repo   *drone.Repo
	Stage  *drone.Stage
	System *drone.System
}
