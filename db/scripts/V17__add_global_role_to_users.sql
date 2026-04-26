-- =============================================
-- Script Name   : V17__add_global_role_to_users.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-26
-- Description   : Add global_role column to users table for SuperAdmin / User distinction
-- =============================================

BEGIN;

ALTER TABLE users
ADD COLUMN IF NOT EXISTS global_role VARCHAR(20) NOT NULL DEFAULT 'User'
CONSTRAINT users_global_role_check CHECK (global_role IN ('SuperAdmin', 'User'));

COMMIT;
