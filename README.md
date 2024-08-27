# goseed

How to use:

Clone the repo

```bash
git clone https://github.com/MatheusLasserre/goseed.git
```

Install dependencies

```bash
cd goseed
go mod tidy
```

Build the binary

```bash
make bi
```

Run the binary

```bash
goseed -d goseed -t person -s 50101 -c 1000 -p "root:goseed@tcp(localhost:3306)/"
```

Use -h for help

```bash
goseed -h
```

