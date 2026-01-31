CREATE table accounts (
    id int not null auto_increment,
    name varchar(255) not null,
    starting_balance int not null,
    created_at timestamp null,
    updated_at timestamp null,

    CONSTRAINT accounts_pk PRIMARY KEY (id)
)

