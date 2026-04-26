-- =============================================
-- Script Name   : V11__create_kitty_cycle_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-20
-- Description   : Creates the kitty_cycle table. A kitty_cycle represents a
--                 multi-month scheduled round of a group where each member
--                 takes a turn hosting and receiving the pool.
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS kitty_cycle (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID NOT NULL,
    name            VARCHAR(255),
    total_cycles    INT  NOT NULL CHECK (total_cycles > 0),
    monthly_amount  NUMERIC(12, 2) NOT NULL CHECK (monthly_amount > 0),
    pool_amount     NUMERIC(12, 2) NOT NULL CHECK (pool_amount > 0),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                        CHECK (status IN ('PENDING', 'ACTIVE', 'COMPLETED', 'CANCELLED')),
    notes           TEXT,
    created_by      UUID NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_kitty_cycle_group_start UNIQUE (group_id, start_date),
    CONSTRAINT fk_kitty_cycle_group   FOREIGN KEY (group_id)   REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_kitty_cycle_creator FOREIGN KEY (created_by) REFERENCES users(id)  ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_kitty_cycle_group  ON kitty_cycle(group_id);
CREATE INDEX IF NOT EXISTS idx_kitty_cycle_status ON kitty_cycle(status);

COMMIT;
