# Database Migrations

Gothic Forge uses [goose](https://github.com/pressly/goose) for database migrations.

## Why Migrations?

Database migrations allow you to:
1. **Version control your database schema** - Track changes over time
2. **Deploy safely** - Apply schema changes incrementally
3. **Rollback when needed** - Undo problematic changes
4. **Collaborate effectively** - Team members stay in sync

## Migration Files

Migrations are SQL files with a specific naming convention:
```
[VERSION]_[DESCRIPTION].sql
```

Examples:
- `00001_initial_schema.sql`
- `00002_add_user_roles.sql`
- `00003_create_comments_table.sql`

## File Structure

Each migration has two sections:

```sql
-- +goose Up
-- SQL to apply the migration
CREATE TABLE users (...);

-- +goose Down
-- SQL to rollback the migration
DROP TABLE users;
```

## Commands

### Apply migrations
```bash
gforge db --migrate
```

### Check migration status
```bash
gforge db --status
```

### Rollback all migrations
```bash
gforge db --reset
```

## Best Practices

1. **Never modify existing migrations** - Create new ones instead
2. **Test rollbacks** - Ensure `Down` sections work correctly
3. **Use transactions** - Most DDL operations in CockroachDB/Postgres are transactional
4. **Add indexes** - Improve query performance
5. **Use UUIDs** - Better for distributed databases like CockroachDB
6. **Document complex migrations** - Add comments explaining WHY

## Educational Resources

- **Goose documentation**: https://github.com/pressly/goose
- **CockroachDB best practices**: https://www.cockroachlabs.com/docs/stable/schema-design-overview
- **PostgreSQL data types**: https://www.postgresql.org/docs/current/datatype.html

## Example: Creating a New Migration

```bash
# Create a new migration file manually
# File: app/db/migrations/00002_add_user_roles.sql

-- +goose Up
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user';
CREATE INDEX idx_users_role ON users(role);

-- +goose Down
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN role;
```

## Automatic Migration Running

Gothic Forge automatically runs migrations after database provisioning during deployment:
1. Database is provisioned (CockroachDB)
2. CONNECTION string is saved to `.env`
3. Migrations are applied automatically
4. Your app is ready to use!

## CockroachDB Specifics

CockroachDB is PostgreSQL-compatible but has some differences:
- **UUID generation**: Use `gen_random_uuid()` (supported)
- **Serial integers**: Use `UNIQUE_ROWID()` or UUIDs instead
- **Transactions**: Fully ACID compliant
- **Indexes**: Automatically distributed across nodes
- **Foreign keys**: Fully supported with CASCADE options

Learn more: https://www.cockroachlabs.com/docs/stable/postgresql-compatibility
