--
DROP DATABASE coffee700;
CREATE DATABASE coffee700 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; 
USE coffee700;
CREATE TABLE users (
    id BIGINT NOT NULL,
    shownName VARCHAR(255) NOT NULL DEFAULT "",
    login VARCHAR(255) NOT NULL,
    status TINYINT NOT NULL DEFAULT 0,
    photoFileId VARCHAR(255) NOT NULL DEFAULT "",
    bio TEXT NOT NULL DEFAULT "",
    createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    lastActiveAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=INNODB;

CREATE TABLE users_settings (
    userId BIGINT,
    settingName VARCHAR(255) NOT NULL,
    settingValue VARCHAR(1024),
    PRIMARY KEY (userId, settingName),
    INDEX fk_userId (userId),
    FOREIGN KEY (userId) 
        REFERENCES users(id)
        ON DELETE CASCADE
) ENGINE=INNODB;

CREATE TABLE users_contacts (
    userId BIGINT,
    contactId BIGINT,
    status TINYINT NOT NULL DEFAULT 0,
    createdAt timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (userId, contactId),
    INDEX fk_userId (userId),
    INDEX fk_contactId (contactId),
    FOREIGN KEY (userId) 
        REFERENCES users(id)
        ON DELETE CASCADE,
    FOREIGN KEY (contactId) 
        REFERENCES users(id)
        ON DELETE CASCADE
) ENGINE=INNODB;
