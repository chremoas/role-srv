CREATE TABLE roles (
  id BIGINT(20) PRIMARY KEY NOT NULL AUTO_INCREMENT,
  name VARCHAR(256) NOT NULL,
  color INT DEFAULT 0,
  hoist BOOL DEFAULT FALSE,
  position INT DEFAULT 1,
  permissions INT DEFAULT 0,
  managed BOOL DEFAULT TRUE,
  mentionable BOOL DEFAULT TRUE,

  role_nick VARCHAR(70),
  inserted TIMESTAMP NOT NULL,
  updated TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX name_uindex ON roles (name);

CREATE TABLE role_membership (
  id BIGINT(20) PRIMARY KEY NOT NULL AUTO_INCREMENT,
  role BIGINT REFERENCES roles (id),
  member VARCHAR(256) NOT NULL
);

CREATE UNIQUE INDEX member_uindex ON role_membership (member);
