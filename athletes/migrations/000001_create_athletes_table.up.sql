CREATE TABLE IF NOT EXISTS athletes (
    first_name varchar(64) NOT NULL,
    last_name varchar(64) NOT NULL,
    start_number integer NOT NULL,
    chip_id uuid PRIMARY KEY
);