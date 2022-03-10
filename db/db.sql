CREATE TABLE request
(
    id       SERIAL PRIMARY KEY,
    method   text NOT NULL,
    scheme   text NOT NULL,
    host     text NOT NULL,
    path     text NOT NULL,
    header   text        default '',
    body     text        default '',
    add_time TIMESTAMPTZ default now()
);

CREATE TABLE response
(
    id           SERIAL PRIMARY KEY,
    req_id       SERIAL REFERENCES request (id) NOT NULL,
    code         INT                            NOT NULL,
    resp_message text                           NOT NULL,
    header       text        default '',
    body         text        default '',
    add_time     TIMESTAMPTZ default now()
);
