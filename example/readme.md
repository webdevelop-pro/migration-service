# Migration usage example
If you haven't read description information yet, please check it [here](../readme.md#database-migration-service)

## Files structure
- `.example.env` - required envs
- `migrations/01_user_users/` - folder with migration for user_users service with priority 1.
- `migrations/01_user_users/01_init.sql` - first user_users migration (create table).
- `migrations/01_user_users/02_add_email.sql` - second user_users migration (add column email to table).
- `migrations/01_user_users/functions/01_email.sql` - user_users migrations for function
- `migrations/01_user_users/seeds/01_seed.sql` - user_users migrations for seeds
- `migrations/01_user_users/views/01_users_with_email.sql` - user_users migrations for view
- `migrations/02_email_emails/` - folder with migration for email_emails service with priority 2. It will be applied after user_users.

Name structure for services and migrations should be `<version>_<description>.sql`. All files except *.sql will be ignored by migrations.
Folders with migration can have any nesting level, but last level should be:
`.../<service>/<optional_folder>/<migration>.sql`
where `optional_folder` can be any word, like `seeds`, `functions`, `views` etc. and shouldn't include underscore character.

**correct structure**:
- `./migrations/some_folder/another_folder/01_user_users/01_init.sql`
- `./migrations/some_folder/another_folder/01_user_users/seeds/01_seed.sql`
- `./migrations/03_email_emails/01_init.sql`
- `./migrations/03_email_emails/views/01_view.sql`

**incorrect structure**:

- `./migrations/email_emails/views/01_view.sql` - `email_emails` should have version
- `./migrations/01_email_emails/all_views/01_view.sql` - `all_views` should include underscore character
- `./migrations/01_email_emails/views/newview/01_view.sql` - service folder can have only 1 nested level


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
*identify last migration id*. We can do this by looking at sql file with last version at `migrations/02_email_emails/` folder, or by doing an sql query `select version from migration_services where name='email_emails'`

*create new file for plus one version* at `migrations/02_email_emails/` folder. For example if last version 2 - we create file with name `03_change_is_canceled_default.sql` where `03` - new version, `change_is_canceled_default` - any string, that describes what you do in migration. 

We need to write valid sql into the file:
```sql
ALTER TABLE public.email_emails ALTER COLUMN is_canceled SET DEFAULT false;
```
Run `./make.sh run` after.

Whooa we fixed email_emails table and created our first migration
