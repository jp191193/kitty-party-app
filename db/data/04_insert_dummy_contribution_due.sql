-- =============================================
-- Script Name   : 04_insert_dummy_contribution_due.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-18
-- Description   : Adds dummy dues for the groups
-- =============================================

BEGIN;

------------------------------------------------------------------
-- 1. Dues for 'Diwali Mega Kitty' (Cycle Month 1)
-- Members: Anjali, Priya
------------------------------------------------------------------
INSERT INTO contribution_due (group_id, member_id, host_member_id, cycle_month, cycle_year, cycle_number, kitty_party_date, due_date, amount, status)
SELECT g.id, u.id, (SELECT id FROM users WHERE email='anjali@example.com'), 1, 2026, 1, '2026-05-10', '2026-05-05', 5000, 'PENDING'
FROM groups g, users u 
WHERE g.name = 'Diwali Mega Kitty' 
  AND u.email IN ('anjali@example.com', 'priya@example.com')
ON CONFLICT (group_id, member_id, cycle_month) DO NOTHING;

------------------------------------------------------------------
-- 2. Dues for 'Summer Vacation Fund' (Cycle Month 1)
-- Members: Rahul, Neha, Anjali
------------------------------------------------------------------
INSERT INTO contribution_due (group_id, member_id, host_member_id, cycle_month, cycle_year, cycle_number, kitty_party_date, due_date, amount, status)
SELECT g.id, u.id, (SELECT id FROM users WHERE email='rahul@example.com'), 1, 2026, 1, '2026-06-15', '2026-06-01', 2500, 'PENDING'
FROM groups g, users u 
WHERE g.name = 'Summer Vacation Fund' 
  AND u.email IN ('rahul@example.com', 'neha@example.com', 'anjali@example.com')
ON CONFLICT (group_id, member_id, cycle_month) DO NOTHING;

COMMIT;
