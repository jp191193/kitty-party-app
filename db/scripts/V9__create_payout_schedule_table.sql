-- =============================================
-- Script Name   : V9__create_payout_schedule_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-19
-- Description   : Authoritative record of who receives the pool each cycle
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS payout_schedule (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id       UUID NOT NULL,
    recipient_id   UUID NOT NULL,
    cycle_number   INT  NOT NULL,
    cycle_month    INT  NOT NULL CHECK (cycle_month BETWEEN 1 AND 12),
    cycle_year     INT  NOT NULL CHECK (cycle_year >= 2000),
    scheduled_date DATE NOT NULL,
    pool_amount    NUMERIC(12, 2) NOT NULL CHECK (pool_amount > 0),
    status         VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED'
                       CHECK (status IN ('SCHEDULED', 'DISBURSED', 'CANCELLED')),
    notes          TEXT,
    created_by     UUID NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_payout_group_cycle UNIQUE (group_id, cycle_number),
    CONSTRAINT fk_payout_group       FOREIGN KEY (group_id)      REFERENCES groups(id)  ON DELETE CASCADE,
    CONSTRAINT fk_payout_recipient   FOREIGN KEY (recipient_id)  REFERENCES users(id)   ON DELETE RESTRICT,
    CONSTRAINT fk_payout_creator     FOREIGN KEY (created_by)    REFERENCES users(id)   ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_payout_group     ON payout_schedule(group_id);
CREATE INDEX IF NOT EXISTS idx_payout_recipient ON payout_schedule(recipient_id);
CREATE INDEX IF NOT EXISTS idx_payout_status    ON payout_schedule(status);

COMMIT;
