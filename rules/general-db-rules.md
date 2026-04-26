# 🗄️ Database Script Rules & Guidelines

## 🎯 Purpose
This document defines the rules and structure for generating and managing database scripts.

All DB scripts must:
- Be **manually executable** (current state)
- Be **pipeline-ready** (future integration)
- Follow **consistent versioning and structure**
- Be **idempotent** (safe to run multiple times)

---

## ⚙️ Current Approach
- Scripts are created and executed manually.
- Developer will copy-paste scripts into DB.
- No CI/CD pipeline is implemented yet.

---

## 🚀 Future Approach (Pipeline Ready)
- Scripts will be automatically executed via pipeline.
- Scripts must not require modification when pipeline is added.
- Execution order will be based on versioning.

---

## 📁 Folder Structure
/db
    /scripts
        V1__init_schema.sql
        V2__create_users_table.sql
        V3__add_indexes.sql

---

## 🏷️ Naming Convention
V{version_number}__{description}.sql


### Examples:
- `V1__init_schema.sql`
- `V2__create_users_table.sql`
- `V3__add_email_index.sql`

---

## 🧱 Standard Script Template

```sql
-- =============================================
-- Script Name   : V{X}__{description}.sql
-- Author        : <Author Name>
-- Created Date  : <YYYY-MM-DD>
-- Description   : <What this script does>
-- =============================================

BEGIN;

-- Example: Create table safely
CREATE TABLE IF NOT EXISTS your_table_name (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMIT;

## Idempotency Rules (MANDATORY)

All scripts MUST:

- Be safe to run multiple times
- Use:
  - IF NOT EXISTS
  - IF EXISTS
- Avoid breaking existing schema