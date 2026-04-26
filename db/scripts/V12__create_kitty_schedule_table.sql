-- =============================================
-- Script Name   : V12__create_kitty_schedule_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-20
-- Description   : Creates the kitty_schedule table. Each row represents one
--                 month within a kitty_cycle: which member hosts/receives
--                 the pool and when the kitty party is held.
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS kitty_schedule (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id        UUID NOT NULL,
    group_id        UUID NOT NULL,
    cycle_number    INT  NOT NULL CHECK (cycle_number > 0),
    cycle_month     INT  NOT NULL CHECK (cycle_month BETWEEN 1 AND 12),
    cycle_year      INT  NOT NULL CHECK (cycle_year >= 2000),
    host_member_id  UUID NOT NULL,
    scheduled_date  DATE NOT NULL,
    due_date        DATE NOT NULL,
    pool_amount     NUMERIC(12, 2) NOT NULL CHECK (pool_amount > 0),
    status          VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED'
                        CHECK (status IN ('SCHEDULED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    notes           TEXT,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_kitty_schedule_cycle_number UNIQUE (cycle_id, cycle_number),
    CONSTRAINT uq_kitty_schedule_cycle_host   UNIQUE (cycle_id, host_member_id),
    CONSTRAINT fk_kitty_schedule_cycle  FOREIGN KEY (cycle_id)       REFERENCES kitty_cycle(id) ON DELETE CASCADE,
    CONSTRAINT fk_kitty_schedule_group  FOREIGN KEY (group_id)       REFERENCES groups(id)      ON DELETE CASCADE,
    CONSTRAINT fk_kitty_schedule_host   FOREIGN KEY (host_member_id) REFERENCES users(id)       ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_kitty_schedule_cycle  ON kitty_schedule(cycle_id);
CREATE INDEX IF NOT EXISTS idx_kitty_schedule_group  ON kitty_schedule(group_id);
CREATE INDEX IF NOT EXISTS idx_kitty_schedule_host   ON kitty_schedule(host_member_id);
CREATE INDEX IF NOT EXISTS idx_kitty_schedule_status ON kitty_schedule(status);

COMMIT;
