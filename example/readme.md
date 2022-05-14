# Migration usage example


## Files structure
- `.example.env` - required envs
- `migrations/01_email.yaml` - first email migration (create table)
- `migrations/functions/01_email.yaml` - second email migration (create trigger)
- `migrations/seeds/01_email.yaml` - third email migration (insert seeds)

## Initialization
- `./make.sh build` - download latest migration binary
- `./make.sh init` - init migration service
- `./make.sh run` - run migration service (dont forget to set up enviroment variables first, `set -a; source .example.env`)

Once you run migration service you should see `email_emails` table with one record in it.

## Usage

### New update for email service
We found an error in our existing table structure for `email` service.
`is_canceled` should be false by default.

In order to make a fix we need to:
*identify last migration id*. We can do this by looking at last version in `migrations/01_email.yaml` file, or by doing an sql query `select version from migration_migration where name='email'`

*create new record with plus one version*. We need to write valid yaml with an sql query
```yaml
- version: 2
  queries:
  - |
    ALTER TABLE public.email_emails ALTER COLUMN is_canceled SET DEFAULT false;
```
and save it in the end of `migrations/01_email.yaml` file. Run `./make.sh run` after.

Whooa we fixed email_emails table and created our first migration
