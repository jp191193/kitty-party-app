-- =============================================
-- Script Name   : R__seed_demo_data.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-21
-- Description   : Repeatable seed that inserts a minimal demo dataset
--                 covering the full kitty cycle lifecycle including the
--                 Start Kitty Cycle flow (PENDING → ACTIVE).
--
--                 Safe to re-run: all inserts use ON CONFLICT DO NOTHING.
--                 Fixed UUIDs allow downstream scripts to reference them.
-- =============================================

BEGIN;

-- ── Members ──────────────────────────────────────────────────────────────────
-- Member 1: group creator / admin
INSERT INTO users (id, name, email, phone, password)
VALUES (
    'a1000000-0000-0000-0000-000000000001',
    'Anjali Sharma',
    'anjali@example.com',
    '9876543210',
    'demo-hashed-password'
)
ON CONFLICT (id) DO NOTHING;

-- Member 2: regular group member
INSERT INTO users (id, name, email, phone, password)
VALUES (
    'a1000000-0000-0000-0000-000000000002',
    'Priya Mehta',
    'priya@example.com',
    '9876543211',
    'demo-hashed-password'
)
ON CONFLICT (id) DO NOTHING;

-- ── Group ─────────────────────────────────────────────────────────────────────
INSERT INTO groups (id, name, monthly_amount, duration, start_date, created_by)
VALUES (
    'b2000000-0000-0000-0000-000000000001',
    'Sharma Kitty Group',
    5000.00,
    12,
    '2026-05-01T00:00:00Z',
    'a1000000-0000-0000-0000-000000000001'
)
ON CONFLICT (id) DO NOTHING;

-- ── Memberships ──────────────────────────────────────────────────────────────
INSERT INTO group_memberships (id, group_id, member_id, status, role)
VALUES (
    'c3000000-0000-0000-0000-000000000001',
    'b2000000-0000-0000-0000-000000000001',
    'a1000000-0000-0000-0000-000000000001',
    'ACTIVE',
    'ADMIN'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO group_memberships (id, group_id, member_id, status, role)
VALUES (
    'c3000000-0000-0000-0000-000000000002',
    'b2000000-0000-0000-0000-000000000001',
    'a1000000-0000-0000-0000-000000000002',
    'ACTIVE',
    'MEMBER'
)
ON CONFLICT (id) DO NOTHING;

-- ── Kitty Cycle (PENDING — ready to be started) ───────────────────────────────
INSERT INTO kitty_cycle (
    id, group_id, name, total_cycles, monthly_amount, pool_amount,
    start_date, end_date, status, notes, created_by
)
VALUES (
    'd4000000-0000-0000-0000-000000000001',
    'b2000000-0000-0000-0000-000000000001',
    '2026 Annual Kitty',
    2,
    5000.00,
    10000.00,
    '2026-05-01',
    '2026-06-30',
    'PENDING',
    'Demo cycle — use PATCH /api/v1/kitty-cycles/{id}/start to activate',
    'a1000000-0000-0000-0000-000000000001'
)
ON CONFLICT (id) DO NOTHING;

-- ── Kitty Schedule entries (SCHEDULED) ────────────────────────────────────────
-- Cycle 1: Anjali hosts in May 2026
INSERT INTO kitty_schedule (
    id, cycle_id, group_id, cycle_number, cycle_month, cycle_year,
    host_member_id, scheduled_date, due_date, pool_amount, status, venue, notes
)
VALUES (
    'e5000000-0000-0000-0000-000000000001',
    'd4000000-0000-0000-0000-000000000001',
    'b2000000-0000-0000-0000-000000000001',
    1, 5, 2026,
    'a1000000-0000-0000-0000-000000000001',
    '2026-05-15',
    '2026-05-10',
    10000.00,
    'SCHEDULED',
    'Anjali''s House',
    'First cycle — will become IN_PROGRESS when cycle is started'
)
ON CONFLICT (id) DO NOTHING;

-- Cycle 2: Priya hosts in June 2026
INSERT INTO kitty_schedule (
    id, cycle_id, group_id, cycle_number, cycle_month, cycle_year,
    host_member_id, scheduled_date, due_date, pool_amount, status, venue, notes
)
VALUES (
    'e5000000-0000-0000-0000-000000000002',
    'd4000000-0000-0000-0000-000000000001',
    'b2000000-0000-0000-0000-000000000001',
    2, 6, 2026,
    'a1000000-0000-0000-0000-000000000002',
    '2026-06-15',
    '2026-06-10',
    10000.00,
    'SCHEDULED',
    'Community Hall',
    'Second cycle'
)
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- ── Quick-reference ───────────────────────────────────────────────────────────
-- Admin member (Anjali)  : a1000000-0000-0000-0000-000000000001
-- Regular member (Priya) : a1000000-0000-0000-0000-000000000002
-- Group                  : b2000000-0000-0000-0000-000000000001
-- Kitty cycle (PENDING)  : d4000000-0000-0000-0000-000000000001
--
-- To test Start Kitty Cycle:
--   1. POST /api/v1/auth/token        { "user_id": "a1000000-0000-0000-0000-000000000001" }
--   2. PATCH /api/v1/kitty-cycles/d4000000-0000-0000-0000-000000000001/start
--      → cycle.status becomes ACTIVE, schedule[0].status becomes IN_PROGRESS
