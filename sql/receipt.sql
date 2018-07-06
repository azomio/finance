CREATE TABLE IF NOT EXISTS receipt (
	fn text,
	i text,
	fp text,
	sum text,
	time integer
);
CREATE UNIQUE INDEX IF NOT EXISTS receipt_ind ON receipt (fn,i,fp);