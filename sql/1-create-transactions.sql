CREATE table transactions (
    id  integer primary key autoincrement not null,
    amount int not NULL,
    date datetime not null,
    description varchar(255) null,
    payee varchar(255) null,
    address varchar(255) null,
    category varchar(255) null,
)

