-- =============================================
-- Script Name   : V14__create_contribution_tracking_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-21
-- Description   : Creates the contribution_tracking table — a per-member
--                 summary per cycle that aggregates total_due, total_paid,
--                 balance and payment status. Depends on groups, users,
--                 and contribution_due.
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS contribution_tracking (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id           UUID NOT NULL,
    member_id          UUID NOT NULL,
    due_id             UUID NOT NULL,
    cycle_month        INT  NOT NULL CHECK (cycle_month BETWEEN 1 AND 12),
    cycle_year         INT  NOT NULL CHECK (cycle_year >= 2000),
    cycle_number       INT  NOT NULL CHECK (cycle_number >= 1),
    total_due          NUMERIC(12,2) NOT NULL CHECK (total_due >= 0),
    total_paid         NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_paid >= 0),
    balance            NUMERIC(12,2) NOT NULL DEFAULT 0,
    status             VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                         CHECK (status IN ('PENDING', 'PARTIAL', 'PAID', 'OVERDUE')),
    last_payment_date  TIMESTAMP NULL,
    created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_tracking_cycle UNIQUE (group_id, member_id, cycle_month, cycle_year),
    CONSTRAINT fk_tracking_group  FOREIGN KEY (group_id)  REFERENCES groups(id)           ON DELETE CASCADE,
    CONSTRAINT fk_tracking_member FOREIGN KEY (member_id) REFERENCES users(id)            ON DELETE CASCADE,
    CONSTRAINT fk_tracking_due    FOREIGN KEY (due_id)    REFERENCES contribution_due(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tracking_group   ON contribution_tracking(group_id);
CREATE INDEX IF NOT EXISTS idx_tracking_member  ON contribution_tracking(member_id);
CREATE INDEX IF NOT EXISTS idx_tracking_due     ON contribution_tracking(due_id);
CREATE INDEX IF NOT EXISTS idx_tracking_status  ON contribution_tracking(status);
CREATE INDEX IF NOT EXISTS idx_tracking_cycle   ON contribution_tracking(cycle_year, cycle_month);

COMMIT;
