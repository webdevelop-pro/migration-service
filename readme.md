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
      created_at timestamp with time zone DEFAULT now() NOT NULL
  );
```


## File structure
Every file represented by `.yaml` standard which that keys:
- `service: <string>` specify service which required this migration
- `migrations: <object>` describe migrations
  - `- version: <int>` set up migration version. Service name + version must be unique
  - `  allowError: boolean`  should migration consider to be successfull even if it failed
  - `  queries: Array<string>` raw SQL for execution

__Example__:
```yaml
service: migration
migrations:
- version: 1
  allowError: false
  queries:
  - |
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
Once github PR reviewed and merged to one of those branches service will execute new migrations automatically


## Env variables
- FORCE_APPLY=true force apply
- APPLY_ONLY=true  only apply
- MIGRATION_DIR=./migrations/  yaml file location
