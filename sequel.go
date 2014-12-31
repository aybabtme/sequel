package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

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

	asPackageFlag := cli.BoolFlag{
		Name:  "pkg",
		Usage: "whether to create a package child of the current directory",
	}

	app.Name = "sequel"
	app.Email = "antoine@do.co"
	app.Author = "Antoine Grondin"
	app.Version = "0.1"
	app.Flags = []cli.Flag{
		usernameFlag,
		passwordFlag,
		dbNameFlag,
		dbAddrFlag,
		asPackageFlag,
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

		var w io.Writer
		if ctx.Bool(asPackageFlag.Name) {
			if err := os.Mkdir(dbname, 0755); err != nil && !os.IsExist(err) {
				log.Fatalf("creating package: %v", err)
			}
			f, err := os.Create(filepath.Join(dbname, dbname+".go"))
			if err != nil {
				log.Fatalf("creating file: %v", err)
			}
			defer f.Close()
			w = f
		} else {
			w = os.Stdout
		}

		if err := generator.Generate(w, schema); err != nil {
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
