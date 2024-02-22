create table buckets (
    id serial primary key,
    name varchar(255) not null unique,
    url varchar(255),
    created_at timestamp not null,
    updated_at timestamp
);


create table users (
    id serial primary key,
    name varchar(255),
    username varchar(255),
    email varchar(255) unique not null,
    password varchar(255) not null,
    inbox int[],
    drives int[],
    friends int[],
    created_at timestamp not null,
    updated_at timestamp
);

create table shares (
    id serial primary key,
    from_user int references users(id) not null,
    to_user int references users(id),
    file int references objects(id),
    note text,
    drive int references drives(id),
    sent timestamp not null
);

create table drives (
    id serial primary key,
    name varchar(255),
    is_personal boolean,
    owner_id int references users(id),
    members int[],
    bucket_id int references buckets(id),
    created_at timestamp not null,
    updated_at timestamp
);

create table objects (
    id serial primary key,
    name varchar(255),
    is_dir boolean,
    drive_id int references drives(id),
    metadata json,
    created_at timestamp not null,
    updated_at timestamp
);