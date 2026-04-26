-- =============================================
-- Script Name   : V10__create_disbursement_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-19
-- Description   : Actual money-transfer record for a payout_schedule entry
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS disbursement (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payout_id       UUID NOT NULL,
    amount_paid     NUMERIC(12, 2) NOT NULL CHECK (amount_paid > 0),
    payment_date    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    payment_mode    VARCHAR(20) NOT NULL CHECK (payment_mode IN ('UPI', 'CASH', 'BANK')),
    transaction_ref VARCHAR(100),
    disbursed_by    UUID NOT NULL,
    notes           TEXT,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_disbursement_payout    FOREIGN KEY (payout_id)    REFERENCES payout_schedule(id) ON DELETE CASCADE,
    CONSTRAINT fk_disbursement_disbursed FOREIGN KEY (disbursed_by) REFERENCES users(id)            ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_disbursement_payout ON disbursement(payout_id);
CREATE INDEX IF NOT EXISTS idx_disbursement_date   ON disbursement(payment_date);

COMMIT;
