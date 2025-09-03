CREATE TABLE IF NOT EXISTS payment_events (
user_id INT NOT NULL,
payment_id INT NOT NULL PRIMARY KEY,
deposit_amount INT NOT NULL
);


CREATE TABLE IF NOT EXISTS skipped_messages (
user_id INT NOT NULL,
payment_id INT NOT NULL PRIMARY KEY,
deposit_amount INT NOT NULL
);