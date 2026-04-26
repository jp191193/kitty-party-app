-- =============================================
-- Script Name   : V1__create_users_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-15
-- Description   : Creates the users table with core fields
--                 for authentication and profile management.
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100)  NOT NULL,
    email       VARCHAR(255)  NOT NULL UNIQUE,
    phone       VARCHAR(15)   NOT NULL UNIQUE,
    password    VARCHAR(255)  NOT NULL,
    is_active   BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
