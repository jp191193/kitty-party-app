-- =============================================
-- Script Name   : V13__add_venue_to_kitty_schedule.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-21
-- Description   : Add optional venue column to kitty_schedule so each
--                 monthly kitty party can record where it was held.
-- =============================================

BEGIN;

ALTER TABLE kitty_schedule
ADD COLUMN IF NOT EXISTS venue TEXT;

COMMIT;
