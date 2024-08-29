DROP DATABASE IF EXISTS goseed;
CREATE DATABASE IF NOT EXISTS goseed;
-- comment
USE goseed;
DROP TABLE IF EXISTS person;
CREATE TABLE IF NOT EXISTS person (
    id BIGINT NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    balance DECIMAL(10,2),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

