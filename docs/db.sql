DROP TABLE IF EXISTS im_user;
CREATE TABLE im_user (
    id SERIAL8 PRIMARY KEY NOT NULL,
    u_id VARCHAR(50) UNIQUE NOT NULL,
    u_name VARCHAR(50) NOT NULL DEFAULT ''
);

DROP TABLE IF EXISTS im_group;
CREATE TABLE im_group (
    id SERIAL8 PRIMARY KEY NOT NULL,
    group_id VARCHAR(50) UNIQUE NOT NULL,
    group_name VARCHAR(50) NOT NULL DEFAULT ''
);

DROP TABLE IF EXISTS im_user_group;
CREATE TABLE im_user_group (
    id SERIAL8 PRIMARY KEY NOT NULL,
    u_id VARCHAR(50) NOT NULL,
    group_id VARCHAR(50) NOT NULL
);

DROP TABLE IF EXISTS im_message_send;
CREATE TABLE im_message_send (
    msg_id SERIAL8 PRIMARY KEY NOT NULL,
    msg_from VARCHAR(50),
    msg_to VARCHAR(50),
    msg_seq BIGINT,
    msg_content VARCHAR(255),
    send_time BIGINT,
    msg_type SMALLINT
);

DROP TABLE IF EXISTS im_message_receive;
CREATE TABLE im_message_receive (
    id SERIAL8 PRIMARY KEY NOT NULL,
    msg_from VARCHAR(50),
    msg_to VARCHAR(50),
    msg_id BIGINT,
    flag SMALLINT DEFAULT 0
);

