Goal: Make this a automatic seed tool for sql databases

Requirements:
- connection string
- database name
- table name
    - Select generator for column



First Draft Goal:
goseed -u database -t table -i quantity -h connection_string


id, name, number

Field: id
Type: int
Null: NO
Key: PRI
Default: NULL
Extra: auto_increment

GenerateMap