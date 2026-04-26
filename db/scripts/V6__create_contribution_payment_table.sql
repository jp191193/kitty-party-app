-- =============================================
-- Script Name   : V6__create_contribution_payment_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-18
-- Description   : Creates the contribution_payment table to record
--                 actual money transactions towards dues.
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS contribution_payment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    due_id          UUID NOT NULL,
    amount_paid     NUMERIC(12,2) NOT NULL CHECK (amount_paid > 0),
    payment_date    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payment_mode    VARCHAR(20) NOT NULL CHECK (payment_mode IN ('UPI', 'CASH', 'BANK')),
    transaction_ref VARCHAR(100),
    status          VARCHAR(20) NOT NULL CHECK (status IN ('SUCCESS', 'FAILED')),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_payment_due FOREIGN KEY (due_id) REFERENCES contribution_due(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_payment_due ON contribution_payment(due_id);
CREATE INDEX IF NOT EXISTS idx_payment_status ON contribution_payment(status);
CREATE INDEX IF NOT EXISTS idx_payment_date ON contribution_payment(payment_date);

COMMIT;
