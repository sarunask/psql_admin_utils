# psql_admin_utils

`psql_admin_utils` is CLI tool for admins of PostgreSQL database. Now it can change owner for database and schemas, user defined types, tables, sequences, views, materialized views, indexes and functions withing that database.

## Install
```
git clone https://github.com/sarunask/psql_admin_utils.git
cd psql_admin_utils
make build
cp ./artifacts/psql_admin_utils /usr/local/bin
```

## Usage

You would need to specify several parameters, similar to `psql` in order to launch:
```
psql_admin_utils chown --host HOST --port 5432 --dbuser postgres --password PASS --database DB --schemas public,your_own_schema
```

You can also use configuration file (default location is $HOME/.psql_admin_utils.yaml), it's simple Yaml file:
```
database: database
host: postgres.host
new-owner: new_owner
password: YOURPASSWORDS
port: 5432
schemas:
- public
tls: false
dbuser: postgres
verbose: true
```
