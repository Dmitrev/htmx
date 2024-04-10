CREATE DATABASE database CHARACTER SET = 'utf8mb4' COLLATE = 'utf8mb4_unicode_ci';
CREATE USER user IDENTIFIED BY 'pass';
GRANT ALL ON `database`.* TO 'user'@'%';
