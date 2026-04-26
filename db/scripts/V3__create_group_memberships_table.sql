-- =============================================
-- Script Name   : V3__create_group_memberships_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-15
-- Description   : Creates the group_memberships table to track members in groups
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS group_memberships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    member_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(group_id, member_id) -- A member can only join a group once
);

COMMIT;
