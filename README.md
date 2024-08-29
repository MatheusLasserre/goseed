# goseed
A CLI tool to seed sql databases with random data.
### How to use:

Clone the repo

```bash
git clone https://github.com/MatheusLasserre/goseed.git && cd goseed
```

Install dependencies

```bash
go mod tidy
```

Build and install the binary

```bash
make bi
```
or

```bash
go build && go install
```

Run the command

```bash
goseed -d mydatabase -t mytable -s 1000000 -c 1000 -p "root:goseed@tcp(localhost:3306)/"
```

Use -h for help

```bash
goseed -h
```
>**Output:**\
>Select a database, a table, and i'll goseed.
>
>Usage:
  goseed [flags]
>
>Flags:
><pre> -c, --chunkSize int     How many rows to insert at a time. Default: 100. Recommended: 10000.\
>  -d, --database string   use database\
>  -h, --help              help for goseed\
>  -p, --host string       Database Connection String. Example: -p "root:goseed@tcp(localhost:3306)/"\
>  --setup-file string       Database Connection String. Example: -p "root:goseed@tcp(localhost:3306)/"\
>  -s, --size int          Seed size\
>  -t, --table string      from table\</pre>

