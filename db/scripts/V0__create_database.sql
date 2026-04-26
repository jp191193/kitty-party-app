-- =============================================
-- Script Name   : V0__create_database.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-15
-- Description   : Creates the initial PostgreSQL database
-- =============================================

-- Note: CREATE DATABASE cannot run inside a transaction block (BEGIN/COMMIT).
-- This script must be executed against the default 'postgres' database 
-- before deploying V1, V2, and V3 to the newly created 'kitty_party' database.

-- Note: When running manually, uncomment the line below. 
-- For Docker, 'docker-compose.yml' inherently creates this database using the POSTGRES_DB variable.
-- CREATE DATABASE kitty_party;

