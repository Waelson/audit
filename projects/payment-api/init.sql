CREATE TABLE IF NOT EXISTS payments (
                                        id SERIAL PRIMARY KEY,
                                        order_number VARCHAR(50) NOT NULL,
                                        payment_amount NUMERIC(10, 2) NOT NULL,
                                        transaction_amount NUMERIC(10, 2) NOT NULL,
                                        name_on_card VARCHAR(255) NOT NULL,
                                        card_number VARCHAR(16) NOT NULL,
                                        expiry_date VARCHAR(5) NOT NULL,
                                        security_code VARCHAR(4) NOT NULL,
                                        postal_code VARCHAR(10) NOT NULL,
                                        transaction_datetime TIMESTAMP NOT NULL
);

ALTER TABLE payments
ALTER COLUMN payment_amount TYPE NUMERIC(15, 2) USING payment_amount::NUMERIC,
ALTER COLUMN transaction_amount TYPE NUMERIC(15, 2) USING transaction_amount::NUMERIC;

CREATE PUBLICATION audit_changes FOR ALL TABLES;