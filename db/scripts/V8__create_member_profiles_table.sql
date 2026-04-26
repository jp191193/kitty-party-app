-- =============================================
-- Script Name   : V8__create_member_profiles_table.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-19
-- Description   : Extended profile information for kitty party members
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS member_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id           UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    date_of_birth       DATE,
    profile_picture_url TEXT,
    address             VARCHAR(255),
    city                VARCHAR(100),
    state               VARCHAR(100),
    pincode             VARCHAR(10),
    occupation          VARCHAR(100),
    about               VARCHAR(500),
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_member_profiles_member_id ON member_profiles(member_id);

COMMIT;
