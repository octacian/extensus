package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/octacian/extensus/core"
	"github.com/octacian/extensus/core/commands"
	"github.com/octacian/migrate"
	"github.com/octacian/shell"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	if os.Getenv("MODE") != "DEV" {
		// Only log the warning severity or above if not in development mode.
		log.SetLevel(log.WarnLevel)
	} else {
		log.WithFields(log.Fields{"MODE": os.Getenv("MODE")}).Info("Development mode enabled")
	}

	// Prepare command-line flags
	flagNoMigrate := flag.Bool("no-migrate", false, "do not apply new migrations")

	flag.Parse() // Parse flags

	// Ensure there are not too many trailing arguments
	if flag.NArg() > 1 {
		log.Panicf("got %d trailing command-line arguments expected 0 to 1: %s", flag.NArg(), os.Args)
	}

	// Defer closing database
	defer core.CloseDB()
	defer core.CloseSQLDB()

	// if the no migrate flag is not true, automatically migrate the database
	if !*flagNoMigrate {
		instance := core.GetMigrate()
		if err := instance.Latest(); err != nil {
			switch err.(type) {
			case *migrate.ErrNoMigrations:
				log.WithFields(log.Fields{"version": instance.Version()}).Info("No database migrations to apply")
			default:
				log.Panic("main: got error while migrating to latest:\n", err)
			}
		}
	}

	// if the trailing argument is equal to shell, launch the shell
	if flag.Arg(0) == "shell" {
		// Register all commands
		commands.Register()
		exitStatus := core.GetShell().Main()
		// Handle exitStatus
		if exitStatus == shell.ExitAll {
			fmt.Printf("Received exit code of %d, exiting...", exitStatus)
		}
	}
}
