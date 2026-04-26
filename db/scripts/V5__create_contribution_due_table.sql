-- =============================================
-- Script Name   : V5__create_contribution_due_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-18
-- Description   : Creates the contribution_due table to manage
--                 monthly financial obligations of members.
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS contribution_due (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL,
    member_id   UUID NOT NULL,
    host_member_id UUID NOT NULL,
    cycle_month INT NOT NULL,
    cycle_year INT NOT NULL,
    cycle_number INT NOT NULL,
    kitty_party_date DATE NOT NULL,
    due_date    DATE NOT NULL,
    amount      NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    status      VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PAID', 'OVERDUE')),
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL,

    CONSTRAINT uq_due UNIQUE (group_id, member_id, cycle_month),
    CONSTRAINT fk_due_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_due_member FOREIGN KEY (member_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_due_group ON contribution_due(group_id);
CREATE INDEX IF NOT EXISTS idx_due_member ON contribution_due(member_id);
CREATE INDEX IF NOT EXISTS idx_due_status ON contribution_due(status);

COMMIT;
