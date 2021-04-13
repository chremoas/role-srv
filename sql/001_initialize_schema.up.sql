--- Filters ------------------------------------

CREATE TABLE filters
(
    id          BIGSERIAL PRIMARY KEY NOT NULL,
    namespace   VARCHAR(32)           NOT NULL,
    name        VARCHAR(32)           NOT NULL,
    description VARCHAR(256)          NOT NULL
);

CREATE UNIQUE INDEX filters_uindex ON filters (name, namespace);

CREATE TABLE filter_membership
(
    id        BIGSERIAL PRIMARY KEY NOT NULL,
    namespace VARCHAR(32)           NOT NULL,
    filter    BIGINT REFERENCES filters (id),
    user_id   BIGINT
);

CREATE UNIQUE INDEX filter_membership_uindex ON filter_membership (filter, user_id, namespace);


--- Roles --------------------------------------

CREATE TABLE roles
(
    id          BIGSERIAL PRIMARY KEY NOT NULL,
    namespace   VARCHAR(32)           NOT NULL,
    color       INT                            DEFAULT 0,
    hoist       BOOL                           DEFAULT FALSE,
    joinable    BOOL                           DEFAULT FALSE,
    managed     BOOL                           DEFAULT TRUE,
    mentionable BOOL                           DEFAULT TRUE,
    name        VARCHAR(256)          NOT NULL,
    permissions INT                            DEFAULT 0,
    position    INT                            DEFAULT 0,
    role_nick   VARCHAR(70)           NOT NULL,
    sig         BOOL                           DEFAULT FALSE,
    sync        BOOL                           DEFAULT FALSE,
    chat_type   VARCHAR(32)           NOT NULL DEFAULT 'discord',
    chat_id     BIGINT                         DEFAULT 0, --- this should probably be `id` but I need to write something to get all the chat ids for that
    inserted    TIMESTAMP             NOT NULL DEFAULT NOW(),
    updated     TIMESTAMP             NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX name_uindex ON roles (name, namespace);

CREATE TABLE role_filters
(
    id        BIGSERIAL PRIMARY KEY NOT NULL,
    namespace VARCHAR(32)           NOT NULL,
    role      BIGINT REFERENCES roles (id),
    filter    BIGINT REFERENCES filters (id)
);

--- Permissions --------------------------------

CREATE TABLE permissions
(
    id          BIGSERIAL PRIMARY KEY NOT NULL,
    namespace   VARCHAR(32)           NOT NULL,
    name        VARCHAR(32)           NOT NULL,
    description VARCHAR(256)          NOT NULL
);

CREATE UNIQUE INDEX permissions_uindex ON permissions (name, namespace);

CREATE TABLE permission_membership
(
    id         BIGSERIAL PRIMARY KEY NOT NULL,
    namespace  VARCHAR(32)           NOT NULL,
    permission BIGINT REFERENCES permissions (id),
    user_id    BIGINT
);

CREATE UNIQUE INDEX permission_membership_uindex ON permission_membership (permission, user_id, namespace);