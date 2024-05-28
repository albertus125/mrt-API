-- init.sql

CREATE TABLE IF NOT EXISTS stations (
    id SERIAL PRIMARY KEY,
    stasiun_name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS schedules (
    id SERIAL PRIMARY KEY,
    station_id INT NOT NULL,
    stasiun_name VARCHAR(255),
    arah VARCHAR(255) NOT NULL,
    jadwal TIME,
    FOREIGN KEY (station_id) REFERENCES stations(id)
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user'
);

CREATE TABLE IF NOT EXISTS reviews (
   id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    rating FLOAT NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id)
);