-- =============================================
-- Script Name   : 02_insert_dummy_groups.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-15
-- Description   : Inserts dummy groups referencing the users table
-- =============================================

BEGIN;

-- Create 'Diwali Mega Kitty' organised by Anjali
INSERT INTO groups (name, monthly_amount, duration, start_date, created_by)
SELECT 'Diwali Mega Kitty', 10000.00, 12, '2026-10-01 00:00:00Z', id 
FROM users 
WHERE email = 'anjali@example.com'
ON CONFLICT DO NOTHING;

-- Create 'Summer Vacation Fund' organised by Rahul
INSERT INTO groups (name, monthly_amount, duration, start_date, created_by)
SELECT 'Summer Vacation Fund', 5000.00, 6, '2026-05-01 00:00:00Z', id 
FROM users 
WHERE email = 'rahul@example.com'
ON CONFLICT DO NOTHING;

COMMIT;
