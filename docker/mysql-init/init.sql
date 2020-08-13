CREATE DATABASE chat;
CREATE USER 'chat_user'@'%' IDENTIFIED WITH mysql_native_password BY 'supersecret';
GRANT SELECT, INSERT, UPDATE, DELETE ON chat.* TO 'chat_user'@'%';

CREATE TABLE chat.subscribers (
account VARCHAR(50) NOT NULL,
chat VARCHAR(15) NOT NULL,
online BOOLEAN NOT NULL DEFAULT TRUE,
banned BOOLEAN NOT NULL DEFAULT FALSE,
subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY(account, chat)
)
