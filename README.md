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

##### _Version v0.8.5_

_Please note that this README describes the current development version and the command line arguments have changed to 
support additional drivers. If you are using v0.8.5, please see [README-v0.8.5](README_v0.8.5.md) for a description of
the command line arguments._

## Releases

| *Version* | *Description*                                                                             |
| --------- | ----------------------------------------------------------------------------------------- |
| v0.8.5    | Initial release with sqlite3 support                                                      |

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

- `load-acl`
- `store-acl`
- `compare-acl`
- `get-acl`
- `put-acl`
- `version`
- `help`

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

### DSN

The `uhppoted-app-db` commands require a DSN command line argument to specify the database connection.

1. For sqlite3, this takes the form `sqlite3://<filepath>`, where the file path is the path to the sqlite3
   database file.
   e.g. `sqlite3://../db/ACL.db`

2. For Microsoft SQL Server, DSN is any DSN accepted by the Microsoft SQL Server driver, as specified in the
   official [documentation](https://pkg.go.dev/github.com/microsoft/go-mssqldb). Typically a SQL Server DSN
   takes the form `sqlserver://<uid>:<password>@<host>?database=<database>`.
   e.g. `sqlserver://sa:UBxNxrQiKWsjncow7mMx@localhost?database=uhppoted`


### `load-acl`

Fetches an ACL file from the configured database and downloads it to the configured UHPPOTE controllers. Intended for use
in a `cron` task that routinely updates the controllers on a scheduled basis.

Command line:

```uhppoted-app-db load-acl```

```uhppoted-app-db  [--debug] [--config <file>] load-acl [--with-pin] --dsn <DSN> [--table:ACL <table>]```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) ACL table. Defaults to ACL.
  --with-pin           Includes the card keypad PIN code when updating the access controllers

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

Command line:

```uhppoted-app-db store-acl```

```uhppoted-app-db [--debug]  [--config <file>] store-acl [--with-pin]  --dsn <DSN> [--table:ACL <table>]```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) ACL table. Defaults to ACL.
  --with-pin           Includes the card keypad PIN code in the information retrieved from the access controllers

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

Command line:

```uhppoted-app-db compare-acl```

```uhppoted-app-db [--debug]  [--config <file>] compare-acl [--with-pin] [--file <file>] --dsn <DSN> [--table:ACL <table>```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) ACL table. Defaults to ACL.
  --with-pin           Includes the card keypad PIN code when comparing card records from  the access controllers
  --file               Optional file path for the compare report. Defaults to displaying the ACL on the console.

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

Command line:

```uhppoted-app-db get-acl``` 

```uhppoted-app-db [--debug] [--config <file>] get-acl [--with-pin] [--file <TSV>] --dsn <DSN> [--table:ACL <table>]```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) ACL table. Defaults to ACL.
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

Command line:

```uhppoted-app-db put-acl --file <TSV> --dsn <DSN>``` 

```uhppoted-app-db [--debug] [--config <file>] put-acl [--with-pin] --file <TSV> --dsn <DSN> [--table:ACL <table>]```

```
  --dsn <DSN>          (required) DSN for database as described above. 
  --table:ACL <table>  (optional) ACL table. Defaults to ACL.
  --with-pin           Includes the card keypad PIN code in the uploaded data
  --file               (required) File path for the TSV file to be uploaded to the database

  --config  Sets the uhppoted.conf file to use for controller configurations
  --debug   Displays verbose debugging information such as the internal structure of the ACL and the
            communications with the UHPPOTE controllers

  Examples:

     uhppoted-app-db put-acl --with-pin --file ACL.tsv --dsn sqlite3://./db/ACL.db --table:ACL ACL2
     uhppoted-app-db --debug --config .uhppoted.conf put-acl --wih-pin --file ACL.tsv --dsn sqlite3://./db/ACL.db
```


