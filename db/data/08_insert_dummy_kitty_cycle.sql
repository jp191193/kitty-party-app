-- =============================================
-- Script Name   : 08_insert_dummy_kitty_cycle.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-20
-- Description   : Dummy kitty_cycle + kitty_schedule rows for existing groups.
--                   Diwali Mega Kitty    → 2 monthly schedule entries
--                   Summer Vacation Fund → 3 monthly schedule entries
-- =============================================

BEGIN;

-- ================================================================
-- GROUP: Diwali Mega Kitty
-- Members : Anjali, Priya   (2 hosts)
-- Organiser / created_by : Anjali
-- ================================================================

INSERT INTO kitty_cycle
    (group_id, name, total_cycles, monthly_amount, pool_amount,
     start_date, end_date, status, notes, created_by)
SELECT
    g.id,
    'Diwali Mega Kitty – 2026 Cycle',
    2,
    10000.00,
    20000.00,
    '2026-10-01'::DATE,
    '2026-11-30'::DATE,
    'ACTIVE',
    'Two-member rotation: Anjali → Priya',
    g.created_by
FROM groups g
WHERE g.name = 'Diwali Mega Kitty'
ON CONFLICT (group_id, start_date) DO NOTHING;

-- Schedule row 1 — Anjali hosts October 2026
INSERT INTO kitty_schedule
    (cycle_id, group_id, cycle_number, cycle_month, cycle_year,
     host_member_id, scheduled_date, due_date, pool_amount, status, notes)
SELECT
    kc.id,
    g.id,
    1, 10, 2026,
    host.id,
    '2026-10-10'::DATE,
    '2026-10-08'::DATE,
    20000.00,
    'COMPLETED',
    'Cycle 1 – Anjali hosts'
FROM kitty_cycle kc
JOIN groups g  ON g.id = kc.group_id
JOIN users host ON host.email = 'anjali@example.com'
WHERE g.name = 'Diwali Mega Kitty'
  AND kc.start_date = '2026-10-01'
ON CONFLICT (cycle_id, cycle_number) DO NOTHING;

-- Schedule row 2 — Priya hosts November 2026
INSERT INTO kitty_schedule
    (cycle_id, group_id, cycle_number, cycle_month, cycle_year,
     host_member_id, scheduled_date, due_date, pool_amount, status, notes)
SELECT
    kc.id,
    g.id,
    2, 11, 2026,
    host.id,
    '2026-11-10'::DATE,
    '2026-11-08'::DATE,
    20000.00,
    'SCHEDULED',
    'Cycle 2 – Priya hosts'
FROM kitty_cycle kc
JOIN groups g  ON g.id = kc.group_id
JOIN users host ON host.email = 'priya@example.com'
WHERE g.name = 'Diwali Mega Kitty'
  AND kc.start_date = '2026-10-01'
ON CONFLICT (cycle_id, cycle_number) DO NOTHING;


-- ================================================================
-- GROUP: Summer Vacation Fund
-- Members : Rahul, Neha, Anjali   (3 hosts)
-- Organiser / created_by : Rahul
-- ================================================================

INSERT INTO kitty_cycle
    (group_id, name, total_cycles, monthly_amount, pool_amount,
     start_date, end_date, status, notes, created_by)
SELECT
    g.id,
    'Summer Vacation Fund – 2026 Cycle',
    3,
    5000.00,
    15000.00,
    '2026-05-01'::DATE,
    '2026-07-31'::DATE,
    'ACTIVE',
    'Three-member rotation: Rahul → Neha → Anjali',
    g.created_by
FROM groups g
WHERE g.name = 'Summer Vacation Fund'
ON CONFLICT (group_id, start_date) DO NOTHING;

-- Schedule row 1 — Rahul hosts May 2026
INSERT INTO kitty_schedule
    (cycle_id, group_id, cycle_number, cycle_month, cycle_year,
     host_member_id, scheduled_date, due_date, pool_amount, status, notes)
SELECT
    kc.id, g.id,
    1, 5, 2026,
    host.id,
    '2026-05-15'::DATE,
    '2026-05-12'::DATE,
    15000.00,
    'COMPLETED',
    'Cycle 1 – Rahul hosts'
FROM kitty_cycle kc
JOIN groups g  ON g.id = kc.group_id
JOIN users host ON host.email = 'rahul@example.com'
WHERE g.name = 'Summer Vacation Fund'
  AND kc.start_date = '2026-05-01'
ON CONFLICT (cycle_id, cycle_number) DO NOTHING;

-- Schedule row 2 — Neha hosts June 2026
INSERT INTO kitty_schedule
    (cycle_id, group_id, cycle_number, cycle_month, cycle_year,
     host_member_id, scheduled_date, due_date, pool_amount, status, notes)
SELECT
    kc.id, g.id,
    2, 6, 2026,
    host.id,
    '2026-06-15'::DATE,
    '2026-06-12'::DATE,
    15000.00,
    'COMPLETED',
    'Cycle 2 – Neha hosts'
FROM kitty_cycle kc
JOIN groups g  ON g.id = kc.group_id
JOIN users host ON host.email = 'neha@example.com'
WHERE g.name = 'Summer Vacation Fund'
  AND kc.start_date = '2026-05-01'
ON CONFLICT (cycle_id, cycle_number) DO NOTHING;

-- Schedule row 3 — Anjali hosts July 2026
INSERT INTO kitty_schedule
    (cycle_id, group_id, cycle_number, cycle_month, cycle_year,
     host_member_id, scheduled_date, due_date, pool_amount, status, notes)
SELECT
    kc.id, g.id,
    3, 7, 2026,
    host.id,
    '2026-07-15'::DATE,
    '2026-07-12'::DATE,
    15000.00,
    'SCHEDULED',
    'Cycle 3 – Anjali hosts'
FROM kitty_cycle kc
JOIN groups g  ON g.id = kc.group_id
JOIN users host ON host.email = 'anjali@example.com'
WHERE g.name = 'Summer Vacation Fund'
  AND kc.start_date = '2026-05-01'
ON CONFLICT (cycle_id, cycle_number) DO NOTHING;

COMMIT;
