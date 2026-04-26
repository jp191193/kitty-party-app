-- =============================================
-- Script Name   : V2__create_groups_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-15
-- Description   : Creates the groups table to store kitty party pools
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS groups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    monthly_amount  NUMERIC(12, 2) NOT NULL CHECK (monthly_amount > 0),
    duration        INT NOT NULL CHECK (duration > 0),
    start_date      TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMIT;
