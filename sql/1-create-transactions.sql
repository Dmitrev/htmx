CREATE table transactions (
    id int not null auto_increment,
    amount int not NULL,
    date datetime not null,
    description varchar(255) null,
    payee varchar(255) null,
    address varchar(255) null,
    category varchar(255) null,
    external_transaction_id varchar(255) null,
    created_at timestamp null,
    updated_at timestamp null,

    CONSTRAINT transactions_pk PRIMARY KEY (id)
)

