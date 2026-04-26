-- =============================================
-- Script Name   : 01_insert_dummy_users.sql
-- Author        : Kitty Party Dev
-- Created Date  : 2026-04-15
-- Description   : Inserts dummy data into the users table 
--                 for local development and testing.
-- =============================================

BEGIN;

INSERT INTO users (id, name, email, phone, password, is_active)
VALUES 
    (gen_random_uuid(), 'Anjali Sharma', 'anjali@example.com', '9876543210', 'hashed_pass_123', true),
    (gen_random_uuid(), 'Priya Mehta', 'priya@example.com', '9123456789', 'hashed_pass_456', true),
    (gen_random_uuid(), 'Rahul Desai', 'rahul@example.com', '9988776655', 'hashed_pass_789', true),
    (gen_random_uuid(), 'Neha Singh', 'neha@example.com', '9876501234', 'hashed_pass_321', true),
    (gen_random_uuid(), 'Vikram Patel', 'vikram@example.com', '9123409876', 'hashed_pass_654', false)
ON CONFLICT (email) DO NOTHING;

COMMIT;
