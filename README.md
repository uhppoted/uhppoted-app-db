![build](https://github.com/uhppoted/uhppoted-app-db/workflows/build/badge.svg)

# uhppoted-app-db

```cron```'able command line utility to download access control lists managed by a database to UHPPOTE
UTO311-L0x access controller boards. 

Supported operating systems:
- Linux
- MacOS
- Windows
- ARM
- ARM7

Supported databases:
- sqlite3
- Microsoft SQL Server
- MySQL
- PostgreSQL

## Release Notes

##### _Version v0.8.5_

1. _Please note that this README describes the current development version and the command line arguments have changed to 
support additional drivers. If you are using v0.8.5, please see [README-v0.8.5](README_v0.8.5.md) for a description of
the command line arguments._

2. The _sqlite3_ implementation has also changed to expect the ACL table to include an ON CONFLICT REPLACE clause when
creating the table rather than implementing an UPSERT.

## Current Release

**[v0.8.9](https://github.com/uhppoted/uhppoted-app-db/releases/tag/v0.8.9) - 2024-09-06**

1. Added TCP/IP support.
2. Updated to Go 1.23.


## Installation

Executables for all the supported operating systems are packaged in the [releases](https://github.com/uhppoted/uhppoted-app-db/releases). The provided archives contain the executables for all the operating systems - OS specific tarballs can be found in the [uhppoted](https://github.com/uhppoted/uhppoted/releases) releases.

Installation is straightforward - download the archive and extract it to a directory of your choice and then place the executable in a directory in your PATH. The `uhppoted-app-db` utility requires the following additional 
files:

- `uhppoted.conf`


### `uhppoted.conf`

`uhppoted.conf` is the communal configuration file shared by all the `uhppoted` project modules and is (or will 
eventually be) documented in [uhppoted](https://github.com/uhppoted/uhppoted). `uhppoted-app-db` requires the 
_devices_ section to resolve non-local controller IP addresses and door to controller door identities.

A sample [uhppoted.conf](https://github.com/uhppoted/uhppoted/blob/master/runtime/simulation/405419896.conf) file is included in the `uhppoted` distribution.

### Building from source

Assuming you have `Go` and `make` installed:

```
git clone https://github.com/uhppoted/uhppoted-app-db.git
cd uhppoted-app-db
make build
```

If you prefer not to use `make`:
```
git clone https://github.com/uhppoted/uhppoted-app-db.git
cd uhppoted-app-db
mkdir bin
go build -trimpath -o bin ./...
```

The above commands build the `'uhppoted-app-db` executable to the `bin` directory.

#### Dependencies

| *Dependency*                                                                 | *Description*                              |
| ---------------------------------------------------------------------------- | ------------------------------------------ |
| [com.github/uhppoted/uhppote-core](https://github.com/uhppoted/uhppote-core) | Device level API implementation            |
| [com.github/uhppoted-lib](https://github.com/uhppoted/uhppoted-lib)          | Shared application library                 |


## uhppoted-app-db

Usage: ```uhppoted-app-db <command> <options>```

Supported commands:

- [`load-acl`](#load-acl)
- [`store-acl`](#store-acl)
- [`compare-acl`](#compare-acl)
- [`get-acl`](#get-acl)
- [`put-acl`](#put-acl)
- [`get-events`](#get-events)
- `version`
- `help`

### DSN

The `uhppoted-app-db` commands require a DSN command line argument to specify the database connection.

1. For sqlite3, this takes the form `sqlite3://<filepath>`, where the file path is the path to the sqlite3
   database file.  
   e.g. `sqlite3://../db/ACL.db`

2. For Microsoft SQL Server, DSN is any DSN accepted by the Microsoft SQL Server driver, as specified in the
   official [documentation](https://pkg.go.dev/github.com/microsoft/go-mssqldb). Typically a SQL Server DSN
   takes the form `sqlserver://<uid>:<password>@<host>?database=<database>`.  
   e.g. `sqlserver://sa:UBxNxrQiKWsjncow7mMx@localhost?database=uhppoted`

3. For MySQL the DSN takes the form `mysql://<DSN>` where the DSN is a [MySQL DSN](https://github.com/go-sql-driver/mysql#dsn-data-source-name), typically of the form `[username[:password]@][protocol[(address)]]/<dbname>`  
   e.g. `mysql://qwerty:uiop@tcp(127.0.0.1:3306)/uhppoted`.
   The database name is required, the other parameters are optional:  
   e.g. a minimal MySQL DSN would be `mysql:///uhppoted` and would use the default connection, user and password to connect
   to database _uhppoted_.

3. For PostgreSQL, DSN is the standard PostgreSQL DSN `postgres://{user}:{password}@{hostname}:{port}/{database-name}`,
   e.g. `postgresql://uhppoted:qwerty@localhost:5432/uhppoted`


### ACL table format

The ACL table is expected to have the following structure:

| Column     | Data Type    | Description                                                                                |
|------------|--------------|--------------------------------------------------------------------------------------------|
| CardNumber | INTEGER      | Valid card number                                                                          |
| PIN        | INTEGER      | Optional keypad PIN code in the range 0-99999. Only required for the --with-pin option.    |
| StartDate  | DATE or TEXT | Date from which the card is valid (YYYY-mm-dd)                                             |
| EndDate    | DATE or TEXT | Date after which the card is no longer valid (YYYY-mm-dd)                                  |
| \<door 1\> | INTEGER      | Access privilege for door 1 (0 none, 1 full access and 2-254 correspond to a time profile) |
| \<door 2\> | INTEGER      | Access privilege for door 2 (0 none, 1 full access and 2-254 correspond to a time profile) |
| ...        | INTEGER      | Access privilege for door N (0 none, 1 full access and 2-254 correspond to a time profile) |

A _Name_ column is optional and ignored.

e.g.:
 
| Name              | CardNumber | PIN   | StartDate  | EndDate    | GreatHall | Gryffindor | HufflePuff | Ravenclaw | Slytherin | Kitchen | Dungeon |Hogsmeade |
|-------------------|------------|-------|------------|------------|-----------|------------|------------|-----------|-----------|---------|---------|----------|
| Albus Dumbledore  | 10058400   | 7531  | 2023-01-01 | 2023-12-31 | 1         | 1          | 1          | 1         | 1         | 1       | 1       | 1        |
| Rubeus Hagrid     | 10058401   | 0     | 2023-01-01 | 2023-12-31 | 1         | 1          | 1          | 1         | 1         | 0       | 0       | 1        |
| Dobby The Elf     | 10058402   | 0     | 2023-01-01 | 2023-12-31 | 1         | 1          | 1          | 1         | 1         | 1       | 0       | 1        |
| Harry Potter      | 10058403   | 0     | 2023-01-01 | 2023-12-31 | 1         | 1          | 0          | 0         | 0         | 0       | 0       | 29       |
| Hermione Grainger | 10058404   | 82953 | 2023-01-01 | 2023-12-31 | 1         | 1          | 0          | 0         | 0         | 0       | 1       | 29       |
| Crookshanks       | 10058405   | 1397  | 2023-01-01 | 2023-12-31 | 0         | 1          | 0          | 0         | 0         | 1       | 0       | 1        |

### Audit trail table format

The audit trail table is optional but if specified on the command line with the`--table:audit` option it is expected to
have the following structure:

| Column     | Data Type    | Description                                                                                |
|------------|--------------|--------------------------------------------------------------------------------------------|
| Timestamp  | DATETIME     | DEFAULT value should be the current date/time                                              |
| Operation  | string       | 'compare', 'load', etc. VARCHAR(64) (or equivalent)                                        |
| Controller | uint32       | Controller ID. INT (or equivalent)                                                         |
| CardNumber | uint32       | Card number. INT (or equivalent)                                                           |
| Status     | string       | Card status. VARCHAR(64) (or equivalent)                                                   |
| Card       | string       | Optional card details. VARCHAR(255) (or equivalent)                                        |

Notes:
1. The table can have either or both of the _CardNumber_ or the _Card_ columns.
2. For sqlite3, SQL Server and MySQL the _Timestamp_ column is expected to be filled automatically with the
   CURRENT_TIMESTAMP.

### Log table format

The operations log table is optional but if specified on the command line with the`--table:log` option it is expected to
have the following structure:

| Column     | Data Type    | Description                                                                                |
|------------|--------------|--------------------------------------------------------------------------------------------|
| Timestamp  | DATETIME     | DEFAULT value should be the current date/time                                              |
| Operation  | string       | 'compare', 'load', etc. VARCHAR(64) (or equivalent)                                        |
| Controller | uint32       | Controller ID. Nullable INT (or equivalent)                                                |
| Detail     | string       | Operation summary. VARCHAR(255) (or equivalent)                                            |

Notes:
1. For sqlite3 and SQL Server the _Timestamp_ column is expected to be filled automatically.


### `load-acl`

Fetches an ACL file from the configured database and downloads it to the configured UHPPOTE controllers. Intended for use
in a `cron` task that routinely updates the controllers on a scheduled basis.

A list of the changes made to the controllers can optionally be stored in an audit trail and a summary of the operation can
optionally be stored in a log table.

Command line:

```uhppoted-app-db load-acl --dsn <DSN>```

```uhppoted-app-db  [--debug] [--config <file>] load-acl [--with-pin] --dsn <DSN> [--table:ACL <table>] [--table:audit <table>] [--table:log <table>]```

```
  --dsn <DSN>            (required) DSN for database as described above. 
  --table:ACL   <table>  (optional) ACL table. Defaults to _ACL_.
  --table:audit <table>  (optional) audit trail table. Defaults to no audit trail.
  --table:log   <table>  (optional) log table. Defaults to no log.
  --with-pin             Includes the card keypad PIN code when updating the access controllers

  --config  Sets the uhppoted.conf file to use for controller configurations
  --debug   Displays verbose debugging information such as the internal structure of the ACL and the
            communications with the UHPPOTE controllers

  Examples:

     uhppoted-app-db load-acl --dsn sqlite3://./db/ACL.db --table:ACL ACL2
     uhppoted-app-db --debug --config .uhppoted.conf load-acl --with-pin --dsn sqlite3://./db/ACL.db
```


### `store-acl`

Fetches the cards stored in the set of configured access controllers, creates a matching ACL from the access controller
configuration and stores it in a database table. Intended for use in a `cron` task that routinely audits the cards stored
on the controllers against an authoritative source. 

A summary of the operation can optionally be appended to stored in a log table.

Command line:

```uhppoted-app-db store-acl --dsn <DSN>```

```uhppoted-app-db [--debug]  [--config <file>] store-acl [--with-pin]  --dsn <DSN> [--table:ACL <table>] [--table:log <table>]```

```
  --dsn <DSN>            (required) DSN for database as described above. 
  --table:ACL   <table>  (optional) ACL table. Defaults to _ACL_.
  --table:audit <table>  (optional) audit trail table. Defaults to no audit trail.
  --table:log   <table>  (optional) log table. Defaults to no log.
  --with-pin             Includes the card keypad PIN code in the information retrieved from the access controllers

  --config  Sets the uhppoted.conf file to use for controller configurations
  --debug   Displays verbose debugging information such as the internal structure of the ACL and the
            communications with the UHPPOTE controllers

  Examples:

     uhppoted-app-db store-acl --dsn sqlite3://./db/ACL.db --table:ACL ACL2
     uhppoted-app-db --debug --config .uhppoted.conf store-acl --with-pin --dsn sqlite3://./db/ACL.db
```


### `compare-acl`

Fetches an ACL file from a database and compares it to the cards stored in the configured UHPPOTE controllers. Intended for 
use in a `cron` task that routinely audits the controllers against an authoritative source.

A list of the differences can optionally be stored in an audit trail and a summary of the operation can optionally be stored
in a log table.

Command line:

```uhppoted-app-db compare-acl --dsn <DSN>```

```uhppoted-app-db [--debug]  [--config <file>] compare-acl [--with-pin] [--file <file>] --dsn <DSN> [--table:ACL <table> [--table:audit <table> [--table:log <table>]```

```
  --dsn <DSN>            (required) DSN for database as described above. 
  --table:ACL <table>    (optional) ACL table. Defaults to _ACL_.
  --table:audit <table>  (optional) audit trail table. Defaults to no audit trail.
  --table:log   <table>  (optional) log table. Defaults to no log.
  --with-pin             Includes the card keypad PIN code when comparing card records from  the access controllers
  --file                 Optional file path for the compare report. Defaults to displaying the ACL on the console.

  --config  Sets the uhppoted.conf file to use for controller configurations
  --debug   Displays verbose debugging information such as the internal structure of the ACL and the
            communications with the UHPPOTE controllers

  Examples:

     uhppoted-app-db compare-acl --dsn sqlite3://./db/ACL.db
     uhppoted-app-db --debug --config .uhppoted.conf compare-acl --with-pin --dsn sqlite3://./db/ACL.db
```


### `get-acl`

Fetches tabular data from a database table and stores it to a TSV file. Intended for use in a `cron` task that routinely
retrieves the ACL from the database for use by scripts on the local host managing the access control system. 

A summary of the operation can optionally be stored in a log table.

Command line:

```uhppoted-app-db get-acl --dsn <DSN>``` 

```uhppoted-app-db [--debug] [--config <file>] get-acl [--with-pin] [--file <TSV>] --dsn <DSN> [--table:ACL <table>] [--table:log <table>]```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) ACL table. Defaults to _ACL_.
  --table:log <table>  (optional) log table. Defaults to no log.
  --with-pin           Includes the card keypad PIN code when retrieving the cards from the access controllers
  --file               Optional file path for the destination TSV file. Defaults to displaying the ACL on
                       the console.
  
  --config  Sets the uhppoted.conf file to use for controller configurations
  --debug   Displays verbose debugging information such as the internal structure of the ACL and the
            communications with the UHPPOTE controllers

  Examples:

     uhppoted-app-db get-acl --dsn sqlite3://./db/ACL.db
     uhppoted-app-db get-acl --dsn sqlite3://./db/ACL.db --table:ACL ACL2 --with-pin
     uhppoted-app-db --debug --config .uhppoted.conf get-acl --dsn sqlite3:./db/ACL.db --with-pin --file ACL.tsv
```

### `put-acl`

Uploads an ACL from a TSV file to a database table. Intended for use in a `cron` task that routinely transfers information
to the database from scripts on the local host.

A summary of the operation can optionally be stored in a log table.

Command line:

```uhppoted-app-db put-acl --file <TSV> --dsn <DSN>``` 

```uhppoted-app-db [--debug] [--config <file>] put-acl [--with-pin] --file <TSV> --dsn <DSN> [--table:ACL <table>] [--table:log <table>]```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) ACL table. Defaults to _ACL_.
  --table:log <table>  (optional) log table. Defaults to no log.
  --with-pin           Includes the card keypad PIN code in the uploaded data
  --file               (required) File path for the TSV file to be uploaded to the database

  --config  Sets the uhppoted.conf file to use for controller configurations
  --debug   Displays verbose debugging information such as the internal structure of the ACL and the
            communications with the UHPPOTE controllers

  Examples:

     uhppoted-app-db put-acl --with-pin --file ACL.tsv --dsn sqlite3://./db/ACL.db --table:ACL ACL2
     uhppoted-app-db --debug --config .uhppoted.conf put-acl --wih-pin --file ACL.tsv --dsn sqlite3://./db/ACL.db
```


### `get-events`

Retrieves events from the set of configured controllers and stores them in a database table, incrementally filling any
gaps in the event list for each controller. 

A summary of the operation can optionally be stored in a log table.

Command line:

```uhppoted-app-db get-events --dsn <DSN>```

```uhppoted-app-db [--debug]  [--config <file>] get-events --dsn <DSN> [--table:events <table>] [--table:log <table>] [--batch-size]```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) Events table. Defaults to _Events_.
  --table:log <table>  (optional) log table. Defaults to no log.
  --batch-size         Maximum number of events to retrieve (per controller) per invocation. Defaults to 128.

  --config  Sets the uhppoted.conf file to use for controller configurations
  --debug   Displays verbose debugging information such as the internal structure of the ACL and the
            communications with the UHPPOTE controllers

  Examples:

     uhppoted-app-db get-events --dsn sqlite3://./db/ACL.db 
     uhppoted-app-db --debug --config .uhppoted.conf get-events --dsn sqlite3://./db/ACL.db --table:events Events2 --batch-size 64
```

