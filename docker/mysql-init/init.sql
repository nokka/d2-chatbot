CREATE DATABASE chat;
CREATE USER 'chat_user'@'%' IDENTIFIED WITH mysql_native_password BY 'supersecret';
GRANT SELECT, INSERT, UPDATE, DELETE ON chat.* TO 'chat_user'@'%';

CREATE TABLE chat.accounts (
id VARCHAR(50) PRIMARY KEY,
online BOOLEAN NOT NULL DEFAULT TRUE,
moderator BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE chat.bans (
account_id VARCHAR(50),
chat VARCHAR(15) NOT NULL,
expires_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY(account_id, chat),
FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE TABLE chat.subscribers (
account_id VARCHAR(50) NOT NULL,
chat VARCHAR(15) NOT NULL,
subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY(account_id, chat),
FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);
