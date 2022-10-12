# Database migration service

## Structure
All migrations files located in the `migrations/` folder.
Migration service reads file one by one in alphabetical order and execute it one by one.
In order to work properly migration service require `migration_service` table to be created first
```sql
CREATE TABLE migration_service (
	id serial NOT NULL PRIMARY KEY,
	name varchar NOT NULL UNIQUE,
	version int NOT NULL DEFAULT 0,
	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone NOT NULL DEFAULT NOW(),
	UNIQUE (name, version)
);
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp_email_emails
  BEFORE UPDATE ON migration_service
  FOR EACH ROW
  EXECUTE PROCEDURE trigger_set_timestamp();
COMMIT;
```
or execute `repositary.CreateMigrationTable` function


## File structure
Every file represented by `.sql` standard which parameters in the first comment.
```sql
- migrations/
- migrations/<PROIRITY>_<service_name>                        --- We set up priority and service name 
- migrations/<PROIRITY>_<service_name>/<VERSION>_<TITLE>.sql  --- We set up migration version and short description
```

__Example__:
```sql
--- allowError: false 
CREATE TABLE user_users(id serial primary key);
CREATE TABLE migration_service (
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
- FORCE_APPLY=true will apply all migration, even market as NoAuto
- APPLY_ONLY=true will only apply transaction but will not start http server
- MIGRATION_DIR=./migrations/  sql file location
