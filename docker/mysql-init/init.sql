CREATE DATABASE chat;
CREATE USER 'chat_user'@'%' IDENTIFIED WITH mysql_native_password BY 'supersecret';
GRANT SELECT, INSERT, UPDATE, DELETE ON chat.* TO 'chat_user'@'%';

CREATE TABLE chat.moderators (
account VARCHAR(50) PRIMARY KEY
);

CREATE TABLE chat.subscribers (
account VARCHAR(50) NOT NULL,
chat VARCHAR(15) NOT NULL,
online BOOLEAN NOT NULL DEFAULT TRUE,
banned_until TIMESTAMP,
subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY(account, chat)
);
