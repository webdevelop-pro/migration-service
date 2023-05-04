# Database migration service


## Structure
All migrations files located in the `migrations/` folder.
Migration service reads file one by one in alphabetical order and execute it one by one.
In order to work properly migration service require `migration_services` table to be created first
```sql
CREATE TABLE IF NOT EXISTS migration_services (
    id serial NOT NULL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    version int NOT NULL DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT NOW()
    );
CREATE OR REPLACE FUNCTION update_at_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER set_timestamp_migration_services
    BEFORE UPDATE ON migration_services
    FOR EACH ROW
    EXECUTE PROCEDURE update_at_set_timestamp();
COMMIT;
```
or execute 
```sh
set -a && source .dev.env && go run cmd/server/main.go --init
```


## File structure
Every file represented by `.sql` standard which parameters in the first comment.
```sql
- migrations/
- migrations/<PROIRITY>_<service_name>                        --- We set up priority and service name 
- migrations/<PROIRITY>_<service_name>/<VERSION>_<TITLE>.sql  --- We set up migration version and short description
```

__Example__:
```sql
--- allow_error: false 
CREATE TABLE user_users(id serial primary key);
CREATE TABLE migration_services (
  id serial NOT NULL PRIMARY KEY,
  name varchar NOT NULL UNIQUE,
  version int NOT NULL DEFAULT 0,
  created_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Usage example
There is two main migration service usage:
- running migrations locally.
```bash
# set -a && source .dev.env && go run cmd/main/main.go
```
will apply all new migrations locally
- automatically applying migrations during merging to dev|stage|master branch
Once github PR reviewed and merged to one of those branches service will execute new migrations automatically.

[Check](example/) full usage example [here](example/)


## Env variables
check `.example.env` file 

## Application options

### --init
creates migration table
```sh 
set -a && source .dev.env && go run cmd/server/main.go --init
```

### --force
force apply migration without version checking. Can accept multiply files or dir paths. Will not update service version if applied version is lower, then already applied
```sh 
set -a && source .dev.env && go run cmd/server/main.go --force ./migrations/01_user_user ./migrations/02_email_emails/02_add_id.sql
```

### --fake
do not apply any migration but mark according migrations in `migration_services` table as completed. Can accept multiply files or dir paths
```sh 
set -a && source .dev.env && go run cmd/server/main.go --fake ./migrations/01_user_user ./migrations/02_email_emails/02_add_id.sql
```

