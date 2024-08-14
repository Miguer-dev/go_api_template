CREATE TABLE IF NOT EXISTS examples (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    example_value_1 numeric(10, 2) NOT NULL DEFAULT 0,
    example_value_2 varchar(4) NOT NULL,
    example_value_3 text NOT NULL
);

ALTER TABLE examples ADD CONSTRAINT example_value_1_check CHECK (example_value_1 >= 0);