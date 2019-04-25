package main

import (
	"flag"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/octacian/extensus/core"
	"github.com/octacian/migrate"
)

func main() {
	// Prepare command-line flags
	flagNoMigrate := flag.Bool("no-migrate", false, "do not apply new migrations")

	flag.Parse() // Parse flags

	// Ensure there are not too many trailing arguments
	if flag.NArg() > 1 {
		panic(fmt.Sprintf("got %d trailing command-line arguments expected 0 to 1", flag.NArg()))
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
				fmt.Printf("Database on version %d with no migrations to apply.\n", instance.Version())
			default:
				panic(fmt.Sprint("main: got error while migrating to latest:\n", err))
			}
		}
	}
}
