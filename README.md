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
goseed -d goseed -t person -s 50101 -c 1000 -p "root:goseed@tcp(localhost:3306)/"
```

Use -h for help

```bash
goseed -h
```

