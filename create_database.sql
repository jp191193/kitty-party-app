-- Create database
CREATE DATABASE kitty_party_db;

-- Connect to the database
\c kitty_party_db

-- Create tables
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE kitty_parties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    organizer_id INTEGER NOT NULL REFERENCES users(id),
    total_amount DECIMAL(10, 2) NOT NULL,
    member_count INTEGER NOT NULL,
    frequency VARCHAR(50),
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE participants (
    id SERIAL PRIMARY KEY,
    party_id INTEGER NOT NULL REFERENCES kitty_parties(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    contribution_amount DECIMAL(10, 2) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(party_id, user_id)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    party_id INTEGER NOT NULL REFERENCES kitty_parties(id),
    payer_id INTEGER NOT NULL REFERENCES users(id),
    receiver_id INTEGER NOT NULL REFERENCES users(id),
    amount DECIMAL(10, 2) NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'pending'
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_kitty_parties_organizer ON kitty_parties(organizer_id);
CREATE INDEX idx_participants_party ON participants(party_id);
CREATE INDEX idx_participants_user ON participants(user_id);
CREATE INDEX idx_transactions_party ON transactions(party_id);