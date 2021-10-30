CREATE TYPE file_part AS (
    storage_url VARCHAR(128),
    file_path VARCHAR(1024),
    content_length BIGINT
    );

CREATE TABLE file
(
    name            VARCHAR(1024) PRIMARY KEY,
    parts           file_part[],
    content_length  BIGINT,
    create_datetime TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE processing_file
(
    name            VARCHAR(1024) PRIMARY KEY,
    parts           file_part[],
    content_length  BIGINT,
    create_datetime TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
