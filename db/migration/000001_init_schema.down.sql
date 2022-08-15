DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS transfers;
/**
Here the droping the 'accounts' is intentionally put in the last
because other two table has Foreign key reference constraints
**/
DROP TABLE IF EXISTS accounts;