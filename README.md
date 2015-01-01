# sequel

Generate a generic Go client to your SQL database.

Only supports MySQL at this time.

## Packages

* `reflector`: connects to a database and inspects its tables and columns.
* `generator`: generates a client package from a `reflector`'s schema.

## todo

Do all the things.

* Be trigger aware.
* Add tests to generated code.
* Create dynamic client from reflector?
* Add support for PostresSQL and SQLite.
