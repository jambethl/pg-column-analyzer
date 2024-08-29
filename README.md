# PostgreSQL Column Analyzer

PostgreSQL Column Analyzer is a CLI tool written in Go which iterates through your PostgreSQL tables and produces a CSV file detailing the column order of each table, reporting how much space is wasted due to inefficient column definition order, helping you optimize your database schemas for ideal performance.

## Index

- [Overview](#overview)
- [Features](#features)
- [Running](#running)
- [Project Structure](#structure)

## Overview
PostgreSQL optimises how data is stored on disk by padding columns so that they align with their neighbouring columns.
For example, a `boolean` column has a size of 1-byte. If we then defined a `smallint` column (2-bytes), an extra byte
would be added to align the 1-byte `boolean` to the 2-byte `smallint`.
Similarly, if we first declared a `integer` column (4-bytes) followed by a `bigint` column (8-bytes), an additional 4-bytes
of padding would be added.

If we declared the above columns as follows:
1. `bigint`
2. `integer`
3. `smallint`
4. `boolean`

There would be no additional padding required.

In a small table, such optimisations may seem unnecessary, but when dealing with billions of records across dozens of tables,
this alignment padding adds up.

## Features
- Connects to your PostgreSQL instance, iterating through its tables in the specified schema
- Analyzes the column order and calculates the space wasted due to inefficient column order definitions
- Generates a CSV report with the optimal column order and space-saving suggestions.

### Example output
| Ordinal Position | Column Name   | Data Type                   | Nullable | Data Type Size (B) | Wasted Padding Per Entry (B) | Recommended Position | Total Wasted Space |
|------------------|---------------|-----------------------------|----------|--------------------|------------------------------|----------------------|--------------------|
| 1                | id            | bigint                      | NO       | 8                  | 0                            | 2                    | 0                  |
| 2                | post_uid      | uuid                        | NO       | 8                  | 0                            | 3                    | 0                  |
| 3                | author_uid    | uuid                        | NO       | 8                  | 2                            | 4                    | 2776               |
| 4                | content       | text                        | NO       | 10                 | 6                            | 1                    | 8328               |
| 5                | created_at    | timestamp without timezone  | NO       | 8                  | 0                            | 5                    | 0                  |
| 6                | like_count    | integer                     | NO       | 4                  | 0                            | 6                    | 0                  |
| 7                | comment_count | integer                     | NO       | 4                  | 0                            | 7                    | 0                  |

## Running
There are five supported optional arguments:
* Database Name
  * name: `database`
  * shorthand: `d`
  * default: `postgres`
* Username
  * name: `username`
  * shorthand: `u`
  * default: `postgres`
* Password
  * name: `password`
  * shorthand: `p`
  * default: `123`
* Host
  * name: `host`
  * shorthand: `l`
  * default: `localhost`
* Schema Name
  * name: `schema`
  * shorthand: `s`
  * default: `public`
* Port
  * name: `port`
  * shorthand: `t`
  * default: `5432`

```sh
go run main.go
```
The above will run the program with the following default values of the above five optional arguments.

To run the program with custom arguments:

```sh
go run main.go -d postgres -u postgres -p 123 -l localhost -s public -t 5432
```

## Structure

### cmd
The `cmd` package contains the code for initialising the CLI with the supported arguments, and retrieving the necessary data from the configured database.

### pkg
The `pkg` contains a few sub-packages as defined below:

* `common` -- contains definitions of structs that are to be shared between files.

* `db` -- responsible for opening the SQL database connection, and can be viewed as an abstraction of the database configuration.

* `report` -- holds the logic for generating the CSV report which informs you of the recommended column order based on data type padding. This is where the supported data typed are defined along with their alignments.
