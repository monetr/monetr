
ALTER TABLE logins ADD COLUMN first_name TEXT;
ALTER TABLE logins ADD COLUMN last_name TEXT;

UPDATE logins SET first_name=users.first_name, last_name=users.last_name
FROM (SELECT users.login_id, users.first_name, users.last_name FROM users) AS users
WHERE logins.login_id = users.login_id;

ALTER TABLE logins ALTER COLUMN first_name SET NOT NULL;