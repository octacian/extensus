package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/octacian/extensus/core"
	"github.com/octacian/shell"
	"golang.org/x/crypto/bcrypt"
)

// Register adds all commands to the shell instance.
func Register() {
	app := core.GetShell()

	app.AddCommand(shell.Command{
		Name:     "migrate",
		Synopsis: "manage database version and migrations",
		Usage: `${name} <sub-command>:

See ${name} help for information on sub-commands.`,
		Main: func(ctx *shell.Context) shell.ExitStatus {
			ctx.App().Println(ctx.Command().Usage)
			return shell.ExitCmd
		},
		SubCommands: []shell.Command{
			{
				Name:     "version",
				Synopsis: "get current database migration version",
				Usage:    "${fullName}",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					ctx.App().Printf("Current migrate version is %d.\n", core.GetMigrate().Version())
					return shell.ExitCmd
				},
			},
			{
				Name:     "latest",
				Synopsis: "migrate database to the latest available version",
				Usage:    "${fullName}",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					if err := core.GetMigrate().Latest(); err != nil {
						ctx.App().Println(err)
					}
					return shell.ExitCmd
				},
			},
			{
				Name:     "to",
				Synopsis: "migrate database to a specific version",
				Usage:    "${fullName} <version number>",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					if ctx.FlagSet().NArg() != 1 {
						ctx.App().Println(ctx.Command().Usage)
						return shell.ExitCmd
					}

					version, err := strconv.Atoi(ctx.FlagSet().Arg(0))
					if err != nil {
						ctx.App().Println(ctx.Command().Usage)
						return shell.ExitCmd
					}

					if err := core.GetMigrate().Goto(version); err != nil {
						ctx.App().Println(err)
					}

					return shell.ExitCmd
				},
			},
			{
				Name:     "list",
				Synopsis: "list all available database migration versions",
				Usage:    "${fullName}",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					ctx.App().Println(core.GetMigrate().List())
					return shell.ExitCmd
				},
			},
		},
	})

	app.AddCommand(shell.Command{
		Name:     "bench-cost",
		Synopsis: "benchmarks how long it takes to generate hashes",
		Usage: `${name} ${shortFlags}:

Benchmark password hash generation with a range of bcrypt cost values. Aim for
a cost that takes around 241 milliseconds per password.

${flags}`,
		SetFlags: func(ctx *shell.Context) {
			flags := ctx.FlagSet()

			ctx.Set("flagStart", flags.Uint("start", uint(bcrypt.MinCost), fmt.Sprintf("Starting cost to test. "+
				"Cannot be greater than end. Cannot be less than %d.", bcrypt.MinCost)))
			ctx.Set("flagEnd", flags.Uint("end", uint(bcrypt.MaxCost), fmt.Sprintf("Ending cost to test. "+
				"Cannot be less than start. Cannot be greater than %d.", bcrypt.MaxCost)))
			ctx.Set("flagTest", flags.String("test", "!test?9@_*", "Plaintext password to test."))
		},
		Main: func(ctx *shell.Context) shell.ExitStatus {
			for cost := *ctx.MustGet("flagStart").(*uint); cost <= *ctx.MustGet("flagEnd").(*uint); cost++ {
				ctx.App().Printf("Cost factor: %d\t\t...", cost)

				start := time.Now()
				_, err := bcrypt.GenerateFromPassword([]byte(*ctx.MustGet("flagTest").(*string)), int(cost))
				if err != nil {
					ctx.App().Printf("Got error while generating hash for cost of %d:\n%s\n", cost, err)
				}
				end := time.Now()

				ctx.App().Printf("\rCost factor: %d\t\t%s\n", cost, end.Sub(start))
			}

			return shell.ExitCmd
		},
	})
}
