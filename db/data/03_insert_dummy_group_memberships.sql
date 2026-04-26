-- =============================================
-- Script Name   : 03_insert_dummy_group_memberships.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-15
-- Description   : Adds dummy members to the groups
-- =============================================

BEGIN;

------------------------------------------------------------------
-- 1. Add Anjali (Organiser) and Priya to 'Diwali Mega Kitty'
------------------------------------------------------------------
INSERT INTO group_memberships (group_id, member_id)
SELECT g.id, u.id 
FROM groups g, users u 
WHERE g.name = 'Diwali Mega Kitty' 
  AND u.email IN ('anjali@example.com', 'priya@example.com')
ON CONFLICT DO NOTHING;

------------------------------------------------------------------
-- 2. Add Rahul (Organiser), Neha, and Anjali to 'Summer Vacation Fund'
------------------------------------------------------------------
INSERT INTO group_memberships (group_id, member_id)
SELECT g.id, u.id 
FROM groups g, users u 
WHERE g.name = 'Summer Vacation Fund' 
  AND u.email IN ('rahul@example.com', 'neha@example.com', 'anjali@example.com')
ON CONFLICT DO NOTHING;

COMMIT;
