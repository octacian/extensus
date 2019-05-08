package commands

import (
	"bufio"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/octacian/extensus/core"
	"github.com/octacian/extensus/core/models"
	"github.com/octacian/shell"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// GetInput reads user input from the shell App's input stream (usually Stdin)
// and returns it as a string. If any errors occur panic is called.
func GetInput(app *shell.App, prefix string) string {
	reader := bufio.NewReader(app.Input)
	app.Printf("%s > ", prefix)
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Panic("GetInput: got error:\n", err)
	}

	return strings.TrimSpace(text)
}

// GetPassword does the same as GetInput but forces the use of Stdin to allow
// preventing input from being echoed back to the terminal as it is entered.
func GetPassword(app *shell.App, prefix string) string {
	app.Printf("%s > ", prefix)
	text, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Panic("GetPassword: got error:\n", err)
	}
	app.Println()

	return strings.TrimSpace(string(text))
}

// getUserByIdentifier takes a user identifier as used by the user get and
// change sub-commands and returns the user referenced or nil if none exist.
// Error messages are printed to the App's output stream.
func getUserByIdentifier(app *shell.App, identifier string) *models.User {
	var user *models.User
	var userErr error

	if identifier[0] == '#' {
		if len(identifier) < 2 {
			app.Println("Expected user ID")
		} else {
			target, err := strconv.Atoi(identifier[1:])
			if err != nil {
				if err, ok := err.(*strconv.NumError); !ok {
					app.Printf("Got unexpected error:\n%s\n", err)
				} else if err.Err == strconv.ErrSyntax {
					app.Printf("Invalid ID number '%s'\n", err.Num)
				} else if err.Err == strconv.ErrRange {
					app.Printf("Number '%s' out of range\n", err.Num)
				} else {
					app.Printf("Got unexpected error with number '%s':\n%s\n", err.Num, err)
				}

				return nil
			}

			user, userErr = models.GetUser(target)
			if userErr != nil {
				if _, ok := userErr.(*models.ErrNoEntry); ok {
					app.Printf("No user with ID %d exists\n", target)
					return nil
				}
			}
		}
	} else {
		user, userErr = models.GetUser(identifier)
		if userErr != nil {
			if _, ok := userErr.(*models.ErrNoEntry); ok {
				app.Printf("No user with email '%s' exists\n", identifier)
				return nil
			}
		}
	}

	if userErr != nil {
		app.Printf("Got unexpected error:\n%s\n", userErr)
		return nil
	}

	return user
}

// checkUserError takes an error and checks if it is a model.ErrInvalid,
// printing the appropriate message to the App's output.
func checkUserError(app *shell.App, err error) {
	if invalid, ok := err.(*models.ErrInvalid); ok {
		switch which := invalid.Which; which {
		case "name", "email":
			app.Printf("Invalid %s '%s'\n", which, invalid.Value)
		case "password":
			app.Println("Invalid password (must be at least 8 characters)")
			app.Printf("Password = '%s'\n", invalid.Value)
		default:
			app.Printf("Got unexpected invalid field %s with value '%s'\n", invalid.Which, invalid.Value)
		}
	} else {
		app.Printf("Got unexpected error:\n%s\n", err)
	}
}

// handleSaveUser takes a user object and attempts to save it, printing the
// appropriate error message depending on the type of error returned.
func handleSaveUser(app *shell.App, user *models.User) {
	if err := user.Save(); err != nil {
		checkUserError(app, err)
	}
}

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

	app.AddCommand(shell.Command{
		Name:     "user",
		Synopsis: "list and manipulate user accounts",
		Usage: `${name} <sub-command>:

	   See user help for more information.`,
		Main: func(ctx *shell.Context) shell.ExitStatus {
			return shell.ExitUsage
		},
		SubCommands: []shell.Command{
			{
				Name:     "list",
				Synopsis: "list user accounts",
				Usage:    "${fullName}",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					users, err := models.ListUser()
					if err != nil {
						if _, ok := err.(*models.ErrEmpty); ok {
							ctx.App().Println("No users exist")
						} else {
							ctx.App().Println(err)
						}
					} else {
						sort.Slice(users, func(i, j int) bool {
							return users[i].Name < users[j].Name
						})

						for _, user := range users {
							ctx.App().Printf("ID:\t\t%d\nName:\t\t%s\nEmail:\t\t%s\nCreated:\t%s\nModified:\t%s\n\n",
								user.ID, user.Name, user.Email, user.Created, user.Modified)
						}
					}

					return shell.ExitCmd
				},
			},
			{
				Name:     "get",
				Synopsis: "show information for a user",
				Usage:    "${fullName} #<user ID>|<user email>",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					if ctx.FlagSet().NArg() != 1 {
						return shell.ExitUsage
					}

					if user := getUserByIdentifier(ctx.App(), ctx.FlagSet().Arg(0)); user != nil {
						ctx.App().Printf("ID:\t\t%d\nName:\t\t%s\nEmail:\t\t%s\nCreated:\t%s\nModified:\t%s\n",
							user.ID, user.Name, user.Email, user.Created, user.Modified)
					}

					return shell.ExitCmd
				},
			},
			{
				Name:     "add",
				Synopsis: "create a new user account",
				Usage:    "${fullName}",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					if ctx.FlagSet().NArg() > 0 {
						return shell.ExitUsage
					}

					name := GetInput(ctx.App(), "Full Name")
					email := GetInput(ctx.App(), "Email")
					password := GetPassword(ctx.App(), "Password")

					user, err := models.NewUser(name, email, password)
					if err != nil {
						checkUserError(ctx.App(), err)
					} else {
						handleSaveUser(ctx.App(), user)
					}

					return shell.ExitCmd
				},
			},
			{
				Name:     "change",
				Synopsis: "change an existing user",
				Usage:    "${fullName} #<user ID>|<user email>",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					if ctx.FlagSet().NArg() != 1 {
						return shell.ExitUsage
					}

					if user := getUserByIdentifier(ctx.App(), ctx.FlagSet().Arg(0)); user != nil {
						ctx.App().Println("Leave input blank to keep values in brackets.")
						name := GetInput(ctx.App(), fmt.Sprintf("Full Name [%s]", user.Name))
						email := GetInput(ctx.App(), fmt.Sprintf("Email [%s]", user.Email))
						password := GetPassword(ctx.App(), "Password [*]")

						if name != "" {
							user.Name = name
						}
						if email != "" {
							user.Email = email
						}
						if password != "" {
							if err := user.SetPassword(password); err != nil {
								checkUserError(ctx.App(), err)
								return shell.ExitCmd
							}
						}

						handleSaveUser(ctx.App(), user)
					}

					return shell.ExitCmd
				},
			},
			{
				Name:     "delete",
				Synopsis: "remove a user",
				Usage:    "${fullName} #<user ID>|<user email>",
				Main: func(ctx *shell.Context) shell.ExitStatus {
					if ctx.FlagSet().NArg() != 1 {
						return shell.ExitUsage
					}

					if user := getUserByIdentifier(ctx.App(), ctx.FlagSet().Arg(0)); user != nil {
						if err := user.Delete(); err != nil {
							ctx.App().Printf("Got unexpected error:\n%s\n", err)
						}
					}

					return shell.ExitCmd
				},
			},
		},
	})
}
