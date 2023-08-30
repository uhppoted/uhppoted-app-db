# CHANGELOG

## Unreleased


## [0.8.6](https://github.com/uhppoted/uhppoted-app-db/releases/tag/v0.8.6) - 2023-08-30

### Added
1. Added support for Microsoft SQL Server.
2. Added optional audit trail for _compare-acl_ and _load-acl_.
3. Added optional operations log for all ACL commands.
4. Implemented `get-events` command to retrieve events from a controller and store
   them in a database table.
5. Added support for MySQL.

### Updated
1. Reworked command line arguments to specify ACL table using --table:ACL option.
2. Reworked sqlite3 implementation to replace the UPSERT with an ON CONFLICT REPLACE
   clause in the CREATE TABLE.
3. Replaced os.Rename with lib implementation for tmpfs support.


## [0.8.5](https://github.com/uhppoted/uhppoted-app-db/releases/tag/v0.8.5) - 2023-06-13

1. Initial release with sqlite3 implementation.