CREATE table subscriptions (
    id uuid primary key default gen_random_uuid(),
    service_name text not null,
    price int not null,
    user_id uuid not null,
    start_date date not null,
    end_date date
);