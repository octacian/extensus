-- @migrate/up
CREATE TABLE blank(
	ID INT PRIMARY KEY
);

-- @migrate/down
DROP TABLE blank;
