CREATE TABLE "dns_log"
(
    "id"               BIGSERIAL PRIMARY KEY,
    "dns_query_record" varchar     NOT NULL,
    "type"             varchar     NOT NULL,
    "ip_address"       varchar     NOT NULL,
    "location"         varchar     NOT NULL,
    "created_at"       timestamptz NOT NULL DEFAULT (now())
);
