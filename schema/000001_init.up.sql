create table if not exists products
(
    id        serial,
    name      varchar unique not null,
    json_data jsonb
);