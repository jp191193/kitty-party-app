-- =============================================
-- Script Name   : V18__expand_group_membership_roles.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-26
-- Description   : Expand group_memberships role values to include TREASURER and VIEWER
-- =============================================

BEGIN;

DO $$
DECLARE
    v_constraint text;
BEGIN
    -- Find the existing CHECK constraint on the role column (added inline in V7)
    SELECT conname INTO v_constraint
    FROM pg_constraint
    WHERE conrelid = 'group_memberships'::regclass
      AND contype = 'c'
      AND pg_get_constraintdef(oid) LIKE '%role%IN%';

    IF v_constraint IS NOT NULL THEN
        EXECUTE 'ALTER TABLE group_memberships DROP CONSTRAINT ' || quote_ident(v_constraint);
    END IF;
END$$;

ALTER TABLE group_memberships
ADD CONSTRAINT group_memberships_role_check
CHECK (role IN ('ADMIN', 'TREASURER', 'MEMBER', 'VIEWER'));

COMMIT;
