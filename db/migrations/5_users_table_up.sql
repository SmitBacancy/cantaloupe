create table users(
	id SERIAL PRIMARY KEY NOT NULL,
	name varchar(256) NOT NULL,
	company_email varchar(32),
	deleted_at timestamp,
	created_at timestamp not null default now(),
	modified_at timestamp
)