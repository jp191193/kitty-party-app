# 📘 ContributionDue Table Design (Kitty App)

## 🎯 Purpose

The `ContributionDue` table represents the **monthly financial obligation** of each member in a group.

It ensures:

* Monthly Finance Obligation is the specific amount of money a member is expected to contribute to the group for a given cycle month.
* Every member has a **trackable due**
* System can manage **pending / paid / overdue**
* No dependency on payment existence

---

## 🧠 Core Principle

> A "due" must exist even if the user never pays.

This table is the **source of truth for obligations**, not transactions.

---

## 🧱 Table Definition (PostgreSQL)

```sql
CREATE TABLE contribution_due (
    id BIGSERIAL PRIMARY KEY,

    group_id BIGINT NOT NULL,
    member_id BIGINT NOT NULL,

    cycle_month INT NOT NULL, -- Month index within kitty cycle (1,2,3...)

    due_date DATE NOT NULL,

    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),

    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    -- Allowed values: PENDING | PAID | OVERDUE

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NULL,

    CONSTRAINT uq_due UNIQUE (group_id, member_id, cycle_month),

    CONSTRAINT fk_due_group FOREIGN KEY (group_id) REFERENCES groups(id),
    CONSTRAINT fk_due_member FOREIGN KEY (member_id) REFERENCES members(id)
);
```

---

## 🔒 Constraints (MANDATORY RULES)

### 1. Unique Constraint

```sql
UNIQUE (group_id, member_id, cycle_month)
```

✅ Prevents duplicate dues for same member and month
❌ Without this → duplicate billing issues

---

### 2. Amount Validation

```sql
CHECK (amount > 0)
```

✅ Ensures no invalid or zero contribution

---

### 3. Foreign Keys

* `group_id → groups(id)`
* `member_id → members(id)`

✅ Maintains relational integrity

---

## 📊 Status Definition

| Status  | Meaning                      |
| ------- | ---------------------------- |
| PENDING | Payment not yet completed    |
| PAID    | Fully paid                   |
| OVERDUE | Due date passed and not paid |

---

## ⚙️ Business Rules

### Rule 1: One Due Per Month

* Each member must have **exactly one due per cycle month**

---

### Rule 2: Due Creation Trigger

Dues should be generated:

* Via API (`GenerateMonthlyDues`)
* OR scheduled job (cron)

---

### Rule 3: Status Handling

* Default → `PENDING`
* If full payment → `PAID`
* If due_date passed and unpaid → `OVERDUE`

---

### Rule 4: No Direct Payment Storage

🚫 Do NOT store:

* amount_paid
* payment_date

👉 Payments belong to `ContributionPayment` table

---

## ⚡ Indexing Strategy

```sql
CREATE INDEX idx_due_group ON contribution_due(group_id);
CREATE INDEX idx_due_member ON contribution_due(member_id);
CREATE INDEX idx_due_status ON contribution_due(status);
```

### Why?

* Fast dashboard queries
* Member contribution lookup
* Pending/overdue filtering

---

## 🔄 Lifecycle Flow

1. Group created
2. Members added
3. System generates dues for each cycle month
4. Member pays → linked via payment table
5. Status updated accordingly

---

## 🧪 Sample Insert

```sql
INSERT INTO contribution_due 
(group_id, member_id, cycle_month, due_date, amount)
VALUES 
(1, 101, 1, '2026-05-05', 5000);
```

---

## 🚨 Important Notes

* Always generate dues **before accepting payments**
* Never delete dues (use soft logic if needed)
* Status should always be **derived via business logic**
* Avoid manual updates from DB — use service layer

---

## 🔮 Future Extensibility

You can later add:

* `late_fee`
* `waived_amount`
* `notes`
* `is_adjusted`

Design supports:

* Partial payments
* AI insights
* Financial reporting

---

## ✅ Summary

| Aspect       | Decision                           |
| ------------ | ---------------------------------- |
| Table Type   | Obligation                         |
| Cardinality  | 1 row per member per month         |
| Critical Key | (group_id, member_id, cycle_month) |
| Dependency   | Independent of payments            |

---