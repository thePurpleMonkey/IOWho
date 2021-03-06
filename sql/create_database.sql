CREATE TABLE IF NOT EXISTS users
(
	user_id SERIAL PRIMARY KEY,
	email VARCHAR(255) UNIQUE NOT NULL,
	password TEXT NOT NULL,
	name VARCHAR(255),
	verified BOOLEAN DEFAULT false,
	restricted BOOLEAN DEFAULT false,
	registered_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	last_login TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verification_emails
(
	user_id INT PRIMARY KEY REFERENCES users(user_id),
	token CHAR(64) NOT NULL,
	sent TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	expires TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP + interval '24 hours' NOT NULL
);

CREATE TABLE IF NOT EXISTS blocked_emails
(
	email TEXT PRIMARY KEY,
	added_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_on TIMESTAMP WITH TIME ZONE,
	deleted_on TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS password_reset
(
	user_id INT PRIMARY KEY REFERENCES users(user_id),
	token CHAR(64) NOT NULL,
	sent TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	expires TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP + interval '24 hours' NOT NULL
);

CREATE TABLE IF NOT EXISTS contacts
(
	contact_id SERIAL PRIMARY KEY,
	user_id INT NOT NULL REFERENCES users(user_id),
	name TEXT UNIQUE NOT NULL,
	email TEXT,
	phone TEXT,
	notes TEXT,
	added_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_on TIMESTAMP WITH TIME ZONE,
	deleted_on TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS transactions
(
	transaction_id SERIAL PRIMARY KEY,
	user_id INT NOT NULL REFERENCES users(user_id),
	contact_id INT REFERENCES contacts(contact_id),
	description TEXT NOT NULL,
	amount MONEY NOT NULL,
	timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
	notes TEXT,
	added_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_on TIMESTAMP WITH TIME ZONE,
	deleted_on TIMESTAMP WITH TIME ZONE
);