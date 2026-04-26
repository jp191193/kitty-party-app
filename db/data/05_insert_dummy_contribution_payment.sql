-- =============================================
-- Script Name   : 05_insert_dummy_contribution_payment.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-18
-- Description   : Adds dummy payments for dues
-- =============================================

BEGIN;

------------------------------------------------------------------
-- 1. Anjali pays full amount (5000) for 'Diwali Mega Kitty'
------------------------------------------------------------------
INSERT INTO contribution_payment (due_id, amount_paid, payment_mode, transaction_ref, status)
SELECT cd.id, 5000, 'UPI', 'UPI_DIWALI_ANJALI1', 'SUCCESS'
FROM contribution_due cd
JOIN groups g ON cd.group_id = g.id
JOIN users u ON cd.member_id = u.id
WHERE g.name = 'Diwali Mega Kitty' 
  AND u.email = 'anjali@example.com'
  AND cd.cycle_month = 1;

-- Also dynamically update the due status for Anjali since she paid the entire amount
UPDATE contribution_due
SET status = 'PAID', updated_at = CURRENT_TIMESTAMP
WHERE id = (
    SELECT cd.id
    FROM contribution_due cd
    JOIN groups g ON cd.group_id = g.id
    JOIN users u ON cd.member_id = u.id
    WHERE g.name = 'Diwali Mega Kitty' 
      AND u.email = 'anjali@example.com'
      AND cd.cycle_month = 1
);

------------------------------------------------------------------
-- 2. Priya makes partial payment (2000) for 'Diwali Mega Kitty'
------------------------------------------------------------------
INSERT INTO contribution_payment (due_id, amount_paid, payment_mode, transaction_ref, status)
SELECT cd.id, 2000, 'BANK', 'BANK_DIWALI_PRIYA1', 'SUCCESS'
FROM contribution_due cd
JOIN groups g ON cd.group_id = g.id
JOIN users u ON cd.member_id = u.id
WHERE g.name = 'Diwali Mega Kitty' 
  AND u.email = 'priya@example.com'
  AND cd.cycle_month = 1;

------------------------------------------------------------------
-- 3. Neha tries to pay for 'Summer Vacation Fund' but fails
------------------------------------------------------------------
INSERT INTO contribution_payment (due_id, amount_paid, payment_mode, transaction_ref, status)
SELECT cd.id, 2500, 'UPI', 'UPI_SUMMER_NEHA_F', 'FAILED'
FROM contribution_due cd
JOIN groups g ON cd.group_id = g.id
JOIN users u ON cd.member_id = u.id
WHERE g.name = 'Summer Vacation Fund' 
  AND u.email = 'neha@example.com'
  AND cd.cycle_month = 1;

COMMIT;
