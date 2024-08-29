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

Current time to generate SQL Value Strings for 100k rows: ~1.87s-1.9s
After changing += to = for tmpValuesString, it took ~1.80s-1.9s
Creating intermediate string for each loop, it took ~660ms

Generating SQL Value Strings for 1M rows: ~17s
After optimizing by making a  tmpSlice instead of a tmpString and just concatenating when the chunkSize is met and then cleaning the tmpArray, it took: ~5s

TODO: Multithreading
    -> Before: 1M rows, 1k chunkSize, it took ~18s
    -> After: 1M rows, 1k chunkSize, it took ~11s
TODO: Add more types
TODO: Support for composite primary keys
TODO: Support for foreign keys
TODO: Make batch insert for each chunkSize