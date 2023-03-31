![build](https://github.com/uhppoted/uhppoted-app-db/workflows/build/badge.svg)

# uhppoted-app-db

```cron```'able command line utility to download access control lists managed by a database to UHPPOTE
UTO311-L0x access controller boards. 

Supported operating systems:
- Linux
- MacOS
- Windows
- ARM7

### Status 

** IN DEVELOPMENT **

## Releases

| *Version* | *Description*                                                                             |
| --------- | ----------------------------------------------------------------------------------------- |
|           |                                                                                           |

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

- `help`
- `version`
- `get-acl`
- `put-acl`
- `load-acl`
- `store-acl`
- `compare-acl`

### ACL table format

### `get-acl`

Fetches tabular data from a database table and stores it to a TSV file. Intended for use in a `cron` task that routinely
retrieves the ACL from the database for use by scripts on the local host managing the access control system. 

Command line:

```uhppoted-app-db get-acl``` 

```uhppoted-app-db [--debug] get-acl [--with-pin] [--file <TSV>]```

```
  --with-pin    Includes the card keypad PIN code in the retrieved file
  --file        Optional file path for the destination TSV file. Defaults to displaying the ACL on the 
                console.
  
  --debug       Displays verbose debugging information such as the internal structure of the ACL and the
                communications with the UHPPOTE controllers
```

### `put-acl`

Uploads an ACL from a TSV file to a database table. Intended for use in a `cron` task that routinely transfers information
to the database from scripts on the local host.

Command line:

```uhppoted-app-db put-acl --file <TSV>``` 

```uhppoted-app-sheets [--debug] put-acl [--with-pin] --file <TSV> [--workdir <dir>]```

```
  --file        File path for the TSV file to be uploaded
  --with-pin    Includes the card keypad PIN code in the uploaded data
  --workdir     Directory for working files, in particular the tokens, revisions, etc
                that provide access to Google Sheets. Defaults to:
                - /var/uhppoted on Linux
                - /usr/local/var/com.github.uhppoted on MacOS
                - ./uhppoted on Microsoft Windows
  --debug       Displays verbose debugging information, in particular the communications
                with the UHPPOTE controllers
```


### `load-acl`

Fetches an ACL file from the configured database and downloads it to the configured UHPPOTE controllers. Intended for use
in a `cron` task that routinely updates the controllers on a scheduled basis.

Command line:

```uhppoted-app-db load-acl```

```uhppoted-app-db load-acl [--debug] [--config <file>] [--no-log] [--no-report] [--workdir <dir>]```

```
  --config      Sets the uhppoted.conf file to use for controller configurations
  --workdir     Sets the working directory for generated report files
  --with-pin    Updates the card keypad PIN code
  --no-log      Writes log messages to the console rather than the rotating log file
  --no-report   Prints the load-acl operational report to the console rather than creating a report file
  --debug       Displays verbose debugging information, in particular the communications with the UHPPOTE controllers
```

### `store-acl`

Fetches the cards stored in the configured UHPPOTE controllers, creates a matching ACL from the UHPPOTED controller configuration and stores it in a database table. Intended for use in a `cron` task that routinely audits the cards stored on the controllers against an authoritative source. 

Command line:

```uhppoted-app-db store-acl```

```uhppoted-app-db store-acl [--debug] [--with-pin] [--no-log] [--config <file>] ```

```
  --config      Sets the uhppoted.conf file to use for controller configurations
  --with-pin    Includes the card keypad PIN code in the retrieved ACL
  --no-log      Writes log messages to the console rather than the rotating log file
  --debug       Displays verbose debugging information, in particular the communications with the UHPPOTE controllers
```

### `compare-acl`

Fetches an ACL file from a database and compares it to the cards stored in the configured UHPPOTE controllers. Intended for 
use in a `cron` task that routinely audits the controllers against an authoritative source.

Command line:

```uhppoted-app-db compare-acl```

```uhppoted-app-db compare-acl [--debug] [-with-pin] [--no-log] [--config <file>]```

```
  --config      Sets the uhppoted.conf file to use for controller configurations
  --with-pin    Includes the card keypad PIN code when comparing cards
  --no-log      Writes log messages to the console rather than the rotating log file
  --debug       Displays verbose debugging information, in particular the communications with the UHPPOTE controllers
```
