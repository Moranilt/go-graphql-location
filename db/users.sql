CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  phone VARCHAR(255),
  email VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CREATE EXTENSION spi;

-- CREATE TRIGGER users_moddatetime
--   BEFORE UPDATE ON users
--   FOR EACH ROW
--   EXECUTE PROCEDURE moddatetime (updated_at);

INSERT INTO users (first_name, last_name, phone, email) VALUES('Name', 'Lastname', '986754673823', 'test@mail.com');
INSERT INTO users (first_name, last_name, phone, email) VALUES('Test', 'TEst', '9878721632163', 'test2@mail.com');