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
    created_at timestamp not null,
    updated_at timestamp
);

create table drives (
    id serial primary key,
    name varchar(255),
    is_personal boolean,
    owner_id int references users(id),
    bucket_id int references buckets(id),
    created_at timestamp not null,
    updated_at timestamp
);

create table drive_members (
    id serial primary key,
    user_id int references users(id) not null,
    drive_id int references drives(id) not null,
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

create table shares (
    id serial primary key,
    from_user int references drive_members(id) not null,
    to_user int references drive_members(id),
    file int references objects(id),
    note text,
    saved boolean,
    sent timestamp not null
);
