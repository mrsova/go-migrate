create table users
(
    id              uuid         not null
        constraint users_id_unique
            unique,
    username        varchar(255)
        constraint users_username_unique
            unique,
);

@DOWN
drop table users;