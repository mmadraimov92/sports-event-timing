#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS athletes (
        first_name varchar(64) NOT NULL,
        last_name varchar(64) NOT NULL,
        start_number integer NOT NULL,
        chip_id uuid PRIMARY KEY
    );
    INSERT INTO athletes (first_name, last_name, start_number, chip_id)
        VALUES ('John', 'Doe', 1, 'd42ebbc6-5b2b-4ff9-83a6-7df87cc20c17'),
                ('Jonah', 'Hubbard', 2, 'e058c321-b904-46ac-a7fb-9bf0ffeb518e'),
                ('Felicia', 'Perez', 3, '32f637d8-40f9-454e-b7b5-88734865cba2'),
                ('Rae', 'Burns', 4, '15c95b2b-e63e-442c-98c4-1be4ac871367');
EOSQL
