package commands

import (
	"fmt"

	flag "github.com/ogier/pflag"
)

// Version is the version associated with this source code.
const Version = "0.0.1"

var version = &Command{
	Run:  versionRun,
	Name: "version",
	ShortUsage: `
USAGE:
    stitch version [--help]
`,
	LongUsage: `Get the version of this CLI.`,
}

var (
	versionFlagSet *flag.FlagSet
)

func init() {
	versionFlagSet = version.initFlags()
}

func versionRun() error {
	if len(versionFlagSet.Args()) > 0 {
		return errUnknownArg(versionFlagSet.Arg(0))
	}
	fmt.Println(Version)
	return nil
}
