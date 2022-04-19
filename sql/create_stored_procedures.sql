-- #region Contact
CREATE OR REPLACE FUNCTION contact_create(
	name TEXT,
	email TEXT,
	phone TEXT,
	notes TEXT
)
	RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE 
-- Variable Declaration
BEGIN
-- Stored Procedure Body
	INSERT INTO contacts (name, email, phone, notes)
	VALUES (name, email, phone, notes)
	RETURNING contact_id;
	-- id := (INSERT INTO contacts (name, email, phone, notes)
	-- VALUES (name, email, phone, notes) RETURNING contact_id);
	-- RETURN id;
END; $$
-- #endregion

-- #region Transaction
CREATE OR REPLACE FUNCTION transaction_create(
	user_id INT,
	description TEXT,
	amount MONEY,
	"timestamp" TIMESTAMP WITH TIME ZONE,
	notes TEXT,
	contact_id INT
)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE 
-- Variable Declaration
BEGIN
-- Stored Procedure Body
	INSERT INTO transactions (user_id, description, amount, timestamp, notes, contact_id)
	VALUES (user_id, description, amount, timestamp, notes, contact_id)
	RETURNING transaction_id;
END; $$

CREATE OR REPLACE FUNCTION transaction_get(
	get_user_id INT,
	get_transaction_id INT
)
RETURNS TABLE (
	user_id INT,
	description TEXT,
	amount MONEY,
	"timestamp" TIMESTAMP WITH TIME ZONE,
	notes TEXT,
	contact_id INT
)
LANGUAGE plpgsql
AS $$
DECLARE 
-- Variable Declaration
BEGIN
-- Stored Procedure Body
	RETURN QUERY SELECT (user_id, description, amount, timestamp, notes, contact_id)
	FROM transactions
	WHERE user_id = get_user_id
	  AND transaction_id = get_transaction_id
	  AND deleted_on IS NULL;
END; $$

CREATE OR REPLACE FUNCTION transaction_update(
	user_id INT,
	transaction_id INT,
	description TEXT,
	amount MONEY,
	"timestamp" TIMESTAMP WITH TIME ZONE,
	notes TEXT,
	contact_id INT
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE 
-- Variable Declaration
BEGIN
-- Stored Procedure Body
	UPDATE transactions 
	SET description = description, amount = amount, timestamp = timestamp, notes = notes, contact_id = contact_id
	WHERE user_id = user_id AND transaction_id = transaction_id;
END; $$

CREATE OR REPLACE FUNCTION transaction_delete(
	user_id INT,
	transaction_id INT
)
RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
DECLARE 
-- Variable Declaration
BEGIN
-- Stored Procedure Body
	WITH rows AS (
		UPDATE transactions 
		SET deleted_on = CURRENT_TIMESTAMP
		WHERE user_id = user_id AND transaction_id = transaction_id
	)
	SELECT COUNT(*) FROM rows;
END; $$
-- #endregion

-- #region Transactions
CREATE OR REPLACE FUNCTION transactions_get_all(
	user_id_parameter int
)
	RETURNS TABLE (
		description TEXT, 
		amount MONEY, 
		"timestamp" TIMESTAMP WITH TIME ZONE,
		notes TEXT 
	)
LANGUAGE plpgsql
AS $$
DECLARE 
-- Variable Declaration
BEGIN
-- Stored Procedure Body
	SELECT t.description, t.amount, t.timestamp, t.notes
	FROM transactions AS t
	WHERE t.user_id = user_id_parameter
	  AND t.deleted_on IS NULL
	ORDER BY t."timestamp" DESC;
END; $$

CREATE OR REPLACE FUNCTION transactions_get_recent(
	user_id_parameter INT,
	record_count INT DEFAULT 10
)
	RETURNS TABLE (
		description TEXT, 
		amount MONEY, 
		"timestamp" TIMESTAMP WITH TIME ZONE,
		notes TEXT 
	)
LANGUAGE plpgsql
AS $$
DECLARE 
-- Variable Declaration
BEGIN
-- Stored Procedure Body
	RETURN QUERY SELECT t.description, t.amount, t.timestamp, t.notes
	FROM transactions AS t
	WHERE t.user_id = user_id_parameter
	  AND t.deleted_on IS NULL
	ORDER BY t."timestamp" DESC
	LIMIT record_count;
END; $$
-- #endregion