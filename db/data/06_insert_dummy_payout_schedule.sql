-- =============================================
-- Script Name   : 06_insert_dummy_payout_schedule.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-19
-- Description   : Dummy payout schedules for existing groups
--                 Diwali Mega Kitty  → 5 cycles  (2 DISBURSED, 2 SCHEDULED, 1 CANCELLED)
--                 Summer Vacation Fund → 4 cycles (2 DISBURSED, 2 SCHEDULED)
-- =============================================

BEGIN;

-- ================================================================
-- GROUP: Diwali Mega Kitty
-- Members : Anjali, Priya  (2 members × ₹10,000 = ₹20,000 pool)
-- Organiser / created_by : Anjali
-- ================================================================

-- Cycle 1 — October 2026 — Anjali receives pool — DISBURSED
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by, updated_at)
SELECT
    g.id,
    recipient.id,
    1, 10, 2026,
    '2026-10-10'::DATE,
    20000.00,
    'DISBURSED',
    'Cycle 1 – Anjali receives the pool',
    g.created_by,
    '2026-10-10 18:00:00+05:30'
FROM groups g
JOIN users recipient ON recipient.email = 'anjali@example.com'
WHERE g.name = 'Diwali Mega Kitty'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

-- Cycle 2 — November 2026 — Priya Mehta receives pool — DISBURSED
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by, updated_at)
SELECT
    g.id,
    recipient.id,
    2, 11, 2026,
    '2026-11-10'::DATE,
    20000.00,
    'DISBURSED',
    'Cycle 2 – Priya Mehta receives the pool',
    g.created_by,
    '2026-11-10 18:00:00+05:30'
FROM groups g
JOIN users recipient ON recipient.email = 'priya@example.com'
WHERE g.name = 'Diwali Mega Kitty'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

-- Cycle 3 — December 2026 — Anjali receives pool — SCHEDULED (upcoming)
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by)
SELECT
    g.id,
    recipient.id,
    3, 12, 2026,
    '2026-12-10'::DATE,
    20000.00,
    'SCHEDULED',
    'Cycle 3 – Anjali scheduled to receive the pool',
    g.created_by
FROM groups g
JOIN users recipient ON recipient.email = 'anjali@example.com'
WHERE g.name = 'Diwali Mega Kitty'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

-- Cycle 4 — January 2027 — Priya Mehta receives pool — SCHEDULED
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by)
SELECT
    g.id,
    recipient.id,
    4, 1, 2027,
    '2027-01-10'::DATE,
    20000.00,
    'SCHEDULED',
    'Cycle 4 – Priya Mehta scheduled to receive the pool',
    g.created_by
FROM groups g
JOIN users recipient ON recipient.email = 'priya@example.com'
WHERE g.name = 'Diwali Mega Kitty'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

-- Cycle 5 — February 2027 — CANCELLED (member withdrew before this cycle)
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by, updated_at)
SELECT
    g.id,
    recipient.id,
    5, 2, 2027,
    '2027-02-10'::DATE,
    20000.00,
    'CANCELLED',
    'Cycle 5 – Cancelled due to group restructuring',
    g.created_by,
    '2027-01-15 10:00:00+05:30'
FROM groups g
JOIN users recipient ON recipient.email = 'anjali@example.com'
WHERE g.name = 'Diwali Mega Kitty'
ON CONFLICT (group_id, cycle_number) DO NOTHING;


-- ================================================================
-- GROUP: Summer Vacation Fund
-- Members : Rahul, Neha, Anjali  (3 members × ₹5,000 = ₹15,000 pool)
-- Organiser / created_by : Rahul
-- ================================================================

-- Cycle 1 — May 2026 — Rahul receives pool — DISBURSED
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by, updated_at)
SELECT
    g.id,
    recipient.id,
    1, 5, 2026,
    '2026-05-15'::DATE,
    15000.00,
    'DISBURSED',
    'Cycle 1 – Rahul receives the pool',
    g.created_by,
    '2026-05-15 17:30:00+05:30'
FROM groups g
JOIN users recipient ON recipient.email = 'rahul@example.com'
WHERE g.name = 'Summer Vacation Fund'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

-- Cycle 2 — June 2026 — Neha receives pool — DISBURSED
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by, updated_at)
SELECT
    g.id,
    recipient.id,
    2, 6, 2026,
    '2026-06-15'::DATE,
    15000.00,
    'DISBURSED',
    'Cycle 2 – Neha receives the pool',
    g.created_by,
    '2026-06-15 17:30:00+05:30'
FROM groups g
JOIN users recipient ON recipient.email = 'neha@example.com'
WHERE g.name = 'Summer Vacation Fund'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

-- Cycle 3 — July 2026 — Anjali receives pool — SCHEDULED
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by)
SELECT
    g.id,
    recipient.id,
    3, 7, 2026,
    '2026-07-15'::DATE,
    15000.00,
    'SCHEDULED',
    'Cycle 3 – Anjali scheduled to receive the pool',
    g.created_by
FROM groups g
JOIN users recipient ON recipient.email = 'anjali@example.com'
WHERE g.name = 'Summer Vacation Fund'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

-- Cycle 4 — August 2026 — Rahul receives pool — SCHEDULED
INSERT INTO payout_schedule
    (group_id, recipient_id, cycle_number, cycle_month, cycle_year,
     scheduled_date, pool_amount, status, notes, created_by)
SELECT
    g.id,
    recipient.id,
    4, 8, 2026,
    '2026-08-15'::DATE,
    15000.00,
    'SCHEDULED',
    'Cycle 4 – Rahul scheduled to receive the pool',
    g.created_by
FROM groups g
JOIN users recipient ON recipient.email = 'rahul@example.com'
WHERE g.name = 'Summer Vacation Fund'
ON CONFLICT (group_id, cycle_number) DO NOTHING;

COMMIT;
