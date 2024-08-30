echo "---------- Benchmark Prog ----------" >> benchmark.txt && \
echo "1M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 1000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "2M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 2000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "3M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 3000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "4M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 4000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "5M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 5000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "6M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 6000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "7M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 7000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "8M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 8000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "9M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 9000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt && \
echo "10M Rows, 10k ChunkSize, " >> benchmark.txt \
    && goseed -d goseed -t person -s 10000000 -c 10000 -p "root:goseed@tcp(localhost:3306)/" --setup-file ../docker/mysql/example.sql \
        | grep "Seed took" >> benchmark.txt && echo ' ' >> benchmark.txt