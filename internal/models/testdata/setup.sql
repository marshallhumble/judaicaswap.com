

create table shares
(
    id          int auto_increment
        primary key,
    owner       int        not null,
    email       text       not null,
    title       text       not null,
    description longtext   not null,
    picture1    text       null,
    picture2    text       null,
    picture3    text       null,
    picture4    text       null,
    picture5    text       null,
    shipsintl   tinyint(1) not null,
    available   tinyint(1) not null,
    createdAt   datetime   null on update CURRENT_TIMESTAMP,
    expires     datetime   null
)
    engine = InnoDB;

create table users
(
    id              int auto_increment
        primary key,
    name            varchar(255)         not null,
    email           varchar(255)         not null,
    hashed_password char(60)             not null,
    created         datetime             not null,
    admin           tinyint(1) default 0 not null,
    user            tinyint(1)           not null,
    guest           tinyint(1)           not null,
    disabled        tinyint(1)           not null,
    Question1       text                 null,
    Question2       text                 null,
    Question3       text                 null,
    constraint users_uc_email
        unique (email)
)
    engine = InnoDB;


INSERT INTO users (name, email, hashed_password, created, admin, user, guest, disabled) VALUES (
     'Alice Jones',
   'alice@example.com',
  '$2a$12$NuTjWXm3KKntReFwyBVHyuf/to.HEwTy.eS206TNfkGfr6HzGJSWG',
 '2022-01-01 09:18:24',
0,
  1,
   0,
   0

                                                                 );
