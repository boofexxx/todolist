CREATE TABLE todolist (
    id bigserial not null primary key,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(50) NOT NULL,
    done BOOLEAN NOT NULL,
    author VARCHAR(50) NOT NULL,
);
