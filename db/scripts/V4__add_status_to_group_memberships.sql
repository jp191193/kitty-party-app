-- =============================================
-- Script Name   : V4__add_status_to_group_memberships.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-16
-- Description   : Add status column to group_memberships table
-- =============================================

BEGIN;

ALTER TABLE group_memberships 
ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' 
CHECK (status IN ('ACTIVE', 'PENDING', 'REJECTED'));

COMMIT;
