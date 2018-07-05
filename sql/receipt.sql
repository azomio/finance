CREATE TABLE IF NOT EXISTS receipt (fn,i,fp,s,t);
CREATE UNIQUE INDEX IF NOT EXISTS receipt_ind ON receipt (fn,i,fp);