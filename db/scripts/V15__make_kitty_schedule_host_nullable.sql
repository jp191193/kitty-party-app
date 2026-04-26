-- =============================================
-- Script Name   : V15__make_kitty_schedule_host_nullable.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-26
-- Description   : Allows host_member_id to be NULL on kitty_schedule rows.
--                 Hosts are now decided by a per-cycle dice roll closer to
--                 the scheduled month rather than chosen up front.
-- =============================================

BEGIN;

ALTER TABLE kitty_schedule
    ALTER COLUMN host_member_id DROP NOT NULL;

COMMIT;
