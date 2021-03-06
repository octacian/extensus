package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	_ "github.com/go-sql-driver/mysql" // Import MySQL database driver
	"github.com/jmoiron/sqlx"
	"github.com/octacian/extensus/shared"
	"github.com/octacian/migrate"
	"github.com/octacian/shell"
	log "github.com/sirupsen/logrus"
)

// Configuration stores a copy of the JSON config file within a native struct.
// Changes made to the struct are not propagated to the file and vise-versa.
type Configuration struct {
	Database struct {
		User     string `json:"user"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	HashCost int    `json:"bcryptCost"`
	Address  string `json:"address"`
	Secret   string `json:"secret"`
}

var sqlDatabase *sql.DB
var oneSQLDatabase sync.Once

var sqlxDatabase *sqlx.DB
var oneSqlxDatabase sync.Once

var migrateInstance *migrate.Instance
var oneMigrateInstance sync.Once

var shellApp *shell.App
var oneShellApp sync.Once

var programConfig Configuration
var oneProgramConfig sync.Once

// GetSQLDB returns a sql.DB for use with packages that do not support sqlx.
func GetSQLDB() *sql.DB {
	oneSQLDatabase.Do(func() {
		config := GetConfig()
		user := config.Database.User
		password := config.Database.Password
		name := config.Database.Name
		res, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?parseTime=true", user, password, name))
		if err != nil {
			log.Panic("GetSQLDB: got error while opening database: ", err)
		}
		sqlDatabase = res
	})

	return sqlDatabase
}

// CloseSQLDB closes the sql.DB instance. It is NOT threadsafe and should only
// be called via defer in main.
func CloseSQLDB() {
	if sqlDatabase != nil {
		sqlDatabase.Close()
		sqlDatabase = nil
		oneSQLDatabase = sync.Once{}
	}
}

// GetDB returns a sqlx.DB.
func GetDB() *sqlx.DB {
	oneSqlxDatabase.Do(func() {
		sqlxDatabase = sqlx.NewDb(GetSQLDB(), "mysql")
		sqlxDatabase.MapperFunc(func(str string) string { return str })
	})

	return sqlxDatabase
}

// CloseDB closes the sqlx.DB instance. It is NOT threadsafe and should only be
// called via defer in main.
func CloseDB() {
	if sqlxDatabase != nil {
		sqlxDatabase.Close()
		sqlxDatabase = nil
		oneSqlxDatabase = sync.Once{}
	}
}

// GetMigrate returns a migrate.Instance.
func GetMigrate() *migrate.Instance {
	oneMigrateInstance.Do(func() {
		result, err := migrate.NewInstance(GetSQLDB(), shared.Abs("migrations"))
		if err != nil {
			log.Panic("GetMigrate: got error while creating instance: ", err)
		}
		migrateInstance = result
	})

	return migrateInstance
}

// GetShell returns a shell.App for use throughout the program
func GetShell() *shell.App {
	oneShellApp.Do(func() {
		shellApp = shell.NewApp("extensus", true)
	})

	return shellApp
}

// GetConfig reads the 'config.json' file at the root of the project and
// returns a struct with its contents. Any fields not defined within the struct
// are ignored.
func GetConfig() *Configuration {
	oneProgramConfig.Do(func() {
		data, err := ioutil.ReadFile(shared.Abs("config.json"))
		if err != nil {
			log.Panic("GetConfig: got error while reading 'config.json': ", err)
		}

		if err := json.Unmarshal(data, &programConfig); err != nil {
			log.Panic("GetConfig: got error while unmarshalling file contests: ", err)
		}
	})

	return &programConfig
}
