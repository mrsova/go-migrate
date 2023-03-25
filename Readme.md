# Go-migrate

Application for creating postgres migration

## Example Usage

### Config structure
```toml
[migrate]
tablename= "migrations"
dir= "./example/migrations"

[database]
username = "user"
password = "pass"
host = "0.0.0.0"
port = "5432"
database = "database"
```

### Example sql file for migrate
@DOWN - separator between up and down
```sql
create table users
(
    id   uuid not null constraint users_id_unique  unique,
    username varchar(255) constraint users_username_unique unique,
);

@DOWN
drop table users;
```

### Run up
```shell
  ./vendor/github.com/mrsova/go-migrate/bin/app --config-path=./example/config/config.toml
```

### Run down decrement by one
```shell
  ./vendor/github.com/mrsova/go-migrate/bin/app --config-path=./example/config/config.toml --rollback=true
```
