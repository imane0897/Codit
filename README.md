# Codit
Elegant Coding Training System

## Pre-requirements

Test on  Ubuntu 16.04.4 x64
-  install Go
```sh
$ sudo apt-get update
$ sudo apt-get install golang-go
```
- install packages
```shell
go get -u github.com/valyala/fasthttp
go get -u github.com/buaazp/fasthttprouter
go get -u github.com/lib/pq
go get -u github.com/satori/go.uuid
go get -u golang.org/x/crypto/bcrypt
```

- install PostgreSQL
```sh
$ sudo apt-get install postgresql postgresql-contrib
```
- Restore SQL

    The installation procedure created a user account called postgres that is associated with the default Postgres role. In order to use Postgres, we can log into that account.
```
// backup old database
pg_dump -U username -h hostname -p port databasename -f filename
// Switch over to the postgres account on your server by typing
$ sudo -i -u postgres
// new database in target sever
$ createdb -T template0 dbname
// add user (in psql)
create role root with login password 'string';
// configure db with dump file
$ psql -U username -h hostname -d desintationdb -p port -f filename
// grant
dbname=# GRANT ALL PRIVILEGES ON TABLE tablename to rolename;
```
