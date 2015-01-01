package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/aybabtme/sequel/generator"
	"github.com/aybabtme/sequel/reflector"
	"github.com/codegangsta/cli"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("sequel: ")
	app := cli.NewApp()

	usernameFlag := cli.StringFlag{
		Name:   "user",
		EnvVar: "MYSQL_USER",
		Value:  "root",
		Usage:  "username to connect to the database",
	}

	passwordFlag := cli.StringFlag{
		Name:   "pass",
		EnvVar: "MYSQL_PASSWORD",
		Value:  "",
		Usage:  "password to connect to the database",
	}

	dbNameFlag := cli.StringFlag{
		Name:   "db",
		EnvVar: "MYSQL_DB_NAME",
		Usage:  "name of the database to connect to",
	}

	dbAddrFlag := cli.StringFlag{
		Name:   "addr",
		EnvVar: "MYSQL_DB_ADDR",
		Value:  "127.0.0.1:3306",
		Usage:  "location of the database to connect to",
	}

	dirFlag := cli.StringFlag{
		Name:  "dir",
		Value: ".",
		Usage: "sub directory where to create the package",
	}

	app.Name = "sequel"
	app.Email = "antoinegrondin@gmail.com"
	app.Author = "Antoine Grondin"
	app.Version = "0.1"
	app.Flags = []cli.Flag{
		usernameFlag,
		passwordFlag,
		dbNameFlag,
		dbAddrFlag,
		dirFlag,
	}
	app.Action = func(ctx *cli.Context) {
		var (
			dsn    string
			dbname = valOrDefault(ctx, dbNameFlag)
		)

		if ctx.IsSet(passwordFlag.Name) {
			dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s",
				valOrDefault(ctx, usernameFlag),
				valOrDefault(ctx, passwordFlag),
				valOrDefault(ctx, dbAddrFlag),
				dbname,
			)
		} else {
			dsn = fmt.Sprintf("%s@tcp(%s)/%s",
				valOrDefault(ctx, usernameFlag),
				valOrDefault(ctx, dbAddrFlag),
				dbname,
			)
		}
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("opening DB: %v", err)
		}
		defer db.Close()

		schema, err := reflector.DescribeMySQL(db, dbname)
		if err != nil {
			log.Fatalf("describing DB: %v", err)
		}

		dirname := valOrDefault(ctx, dirFlag)
		if err := generator.Generate(dirname, schema); err != nil {
			log.Fatalf("generating schema: %v", err)
		}
	}

	app.Run(os.Args)
}

func valOrDefault(ctx *cli.Context, f cli.StringFlag) string {
	str := ctx.String(f.Name)
	if str != "" {
		return str
	}
	if f.Value == "" {
		log.Fatalf("flag not set: %q", f.Name)
	}
	return f.Value
}
