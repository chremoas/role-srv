DROP INDEX role_name_uindex ON roles;
DROP TABLE roles;
DROP INDEX member_uindex ON role_membership;
CREATE TABLE role_membership;
