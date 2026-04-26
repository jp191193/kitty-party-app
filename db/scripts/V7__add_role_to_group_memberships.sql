-- =============================================
-- Script Name   : V7__add_role_to_group_memberships.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-19
-- Description   : Add role column to group_memberships table
-- =============================================

BEGIN;

ALTER TABLE group_memberships 
ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'MEMBER' 
CHECK (role IN ('ADMIN', 'MEMBER'));

-- Update existing memberships where the member is also the creator of the group to be 'ADMIN'
UPDATE group_memberships gm
SET role = 'ADMIN'
FROM groups g
WHERE gm.group_id = g.id AND gm.member_id = g.created_by;

COMMIT;
