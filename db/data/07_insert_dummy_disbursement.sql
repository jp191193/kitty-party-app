-- =============================================
-- Script Name   : 07_insert_dummy_disbursement.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-19
-- Description   : Disbursement records for the DISBURSED payout cycles
--                 Must run AFTER 06_insert_dummy_payout_schedule.sql
-- =============================================

BEGIN;

-- ================================================================
-- GROUP: Diwali Mega Kitty — DISBURSED payouts (Cycles 1 & 2)
-- ================================================================

-- Disbursement for Cycle 1 — ₹20,000 paid to Anjali via UPI
INSERT INTO disbursement
    (payout_id, amount_paid, payment_date, payment_mode, transaction_ref, disbursed_by, notes)
SELECT
    ps.id,
    20000.00,
    '2026-10-10 18:00:00+05:30',
    'UPI',
    'TXN-DWL-OCT2026-001',
    g.created_by,
    'October 2026 pool disbursed to Anjali via UPI'
FROM payout_schedule ps
JOIN groups g ON ps.group_id = g.id
WHERE g.name = 'Diwali Mega Kitty'
  AND ps.cycle_number = 1
  AND NOT EXISTS (SELECT 1 FROM disbursement d WHERE d.payout_id = ps.id);

-- Disbursement for Cycle 2 — ₹20,000 paid to Priya via Bank Transfer
INSERT INTO disbursement
    (payout_id, amount_paid, payment_date, payment_mode, transaction_ref, disbursed_by, notes)
SELECT
    ps.id,
    20000.00,
    '2026-11-10 18:00:00+05:30',
    'BANK',
    'BANK-DWL-NOV2026-002',
    g.created_by,
    'November 2026 pool disbursed to Priya via bank transfer'
FROM payout_schedule ps
JOIN groups g ON ps.group_id = g.id
WHERE g.name = 'Diwali Mega Kitty'
  AND ps.cycle_number = 2
  AND NOT EXISTS (SELECT 1 FROM disbursement d WHERE d.payout_id = ps.id);


-- ================================================================
-- GROUP: Summer Vacation Fund — DISBURSED payouts (Cycles 1 & 2)
-- ================================================================

-- Disbursement for Cycle 1 — ₹15,000 paid to Rahul via UPI
INSERT INTO disbursement
    (payout_id, amount_paid, payment_date, payment_mode, transaction_ref, disbursed_by, notes)
SELECT
    ps.id,
    15000.00,
    '2026-05-15 17:30:00+05:30',
    'UPI',
    'TXN-SUM-MAY2026-001',
    g.created_by,
    'May 2026 pool disbursed to Rahul via UPI'
FROM payout_schedule ps
JOIN groups g ON ps.group_id = g.id
WHERE g.name = 'Summer Vacation Fund'
  AND ps.cycle_number = 1
  AND NOT EXISTS (SELECT 1 FROM disbursement d WHERE d.payout_id = ps.id);

-- Disbursement for Cycle 2 — ₹15,000 paid to Neha via Cash
INSERT INTO disbursement
    (payout_id, amount_paid, payment_date, payment_mode, transaction_ref, disbursed_by, notes)
SELECT
    ps.id,
    15000.00,
    '2026-06-15 17:30:00+05:30',
    'CASH',
    NULL,
    g.created_by,
    'June 2026 pool handed over to Neha in cash at the party'
FROM payout_schedule ps
JOIN groups g ON ps.group_id = g.id
WHERE g.name = 'Summer Vacation Fund'
  AND ps.cycle_number = 2
  AND NOT EXISTS (SELECT 1 FROM disbursement d WHERE d.payout_id = ps.id);

COMMIT;
