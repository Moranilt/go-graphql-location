CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  login VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  phone VARCHAR(255) UNIQUE,
  email VARCHAR(255) UNIQUE NOT NULL,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CREATE EXTENSION spi;

-- CREATE TRIGGER users_moddatetime
--   BEFORE UPDATE ON users
--   FOR EACH ROW
--   EXECUTE PROCEDURE moddatetime (updated_at);

INSERT INTO users (first_name, last_name, phone, login, password, email) VALUES('Name', 'Lastname', '986754673823', 'name', '325325326', 'test@mail.com');
INSERT INTO users (first_name, last_name, phone, login, password, email) VALUES('Test', 'TEst', '9878721632163', 'test', '24124', 'test2@mail.com');