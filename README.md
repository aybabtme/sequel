# sequel

Generate a generic Go client to your SQL database.

Only supports MySQL at this time.

## Tool

The root of this repository is a CLI tool, install it the normal way:

```
go get github.com/aybabtme/sequel
```

To use it:

```bash
$ sequel --user 'root' \
         --pass 'maybe_use_an_env_var?' \
         --db   'my_database' \
         --addr '127.0.0.1:3306' \
         --dir  './in/this/subdir'
```

You can also use env vars for the database details:

```bash
$ export MYSQL_USER="super_granted_user"
$ export MYSQL_PASSWORD="super_secret_password"
$ export MYSQL_DB_NAME="super_well_designed_db"
$ export MYSQL_DB_ADDR="127.0.0.1:3306"
$ sequel
```

## Packages

* `reflector`: connects to a database and inspects its tables and columns.
* `generator`: generates a client package from a `reflector`'s schema.

## todo

Do all the things.

* Add tests to generated table code.
* Be triggers aware.
* Be constraints aware.
* Take row types as arguments for CRUD and List when their ID is expected.
* Create dynamic client from reflector?
* Add support for PostresSQL and SQLite.
