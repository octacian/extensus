package commands

import (
	"strconv"

	"github.com/octacian/extensus/core"
	"github.com/octacian/shell"
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
}
