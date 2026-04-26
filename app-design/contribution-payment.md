# 📘 ContributionPayment Table Design (Kitty App)

## 🎯 Purpose

The `ContributionPayment` table represents **actual money transactions** made by members toward their dues.

It ensures:

* Every payment is **recorded with full audit trail**
* Supports **partial payments, retries, failures**
* Enables financial transparency and dispute handling

---

## 🧠 Core Principle

> A payment is an event, not a state.

This table captures **what actually happened**, while `ContributionDue` tracks **what should happen**.

---

## 🧱 Table Definition (PostgreSQL)

```sql id="x8f2la"
CREATE TABLE contribution_payment (
    id BIGSERIAL PRIMARY KEY,

    due_id BIGINT NOT NULL,

    amount_paid NUMERIC(12,2) NOT NULL CHECK (amount_paid > 0),

    payment_date TIMESTAMP NOT NULL DEFAULT NOW(),

    payment_mode VARCHAR(20) NOT NULL,
    -- Allowed values: UPI | CASH | BANK

    transaction_ref VARCHAR(100),
    -- UPI reference / bank transaction ID

    status VARCHAR(20) NOT NULL,
    -- Allowed values: SUCCESS | FAILED

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payment_due 
        FOREIGN KEY (due_id) REFERENCES contribution_due(id)
);
```

---

## 🔒 Constraints (MANDATORY RULES)

### 1. Amount Validation

```sql id="n1k9qz"
CHECK (amount_paid > 0)
```

✅ Prevents invalid transactions

---

### 2. Foreign Key Integrity

```sql id="m8t3yx"
FOREIGN KEY (due_id) REFERENCES contribution_due(id)
```

✅ Ensures every payment is linked to a valid due

---

### 3. No Unique Constraint on Payments

❗ IMPORTANT:

Do NOT enforce uniqueness on `due_id`

👉 Reason:

* Multiple payments allowed (partial payments)
* Retry scenarios must be recorded

---

## 📊 Status Definition

| Status  | Meaning                        |
| ------- | ------------------------------ |
| SUCCESS | Payment completed successfully |
| FAILED  | Payment attempt failed         |

---

## ⚙️ Business Rules

### Rule 1: Payments Are Append-Only

* Never update or overwrite a payment
* Always insert a new record

---

### Rule 2: Only SUCCESS Counts Financially

* While calculating totals:

```sql id="3s8k2n"
WHERE status = 'SUCCESS'
```

---

### Rule 3: Payment Does NOT Automatically Mean Paid

* After inserting payment:

  * Recalculate total paid
  * Update `ContributionDue.status`

---

### Rule 4: Transaction Reference Handling

* Optional for CASH
* Mandatory for UPI/BANK (recommended)

---

## ⚡ Indexing Strategy

```sql id="7a2kpl"
CREATE INDEX idx_payment_due ON contribution_payment(due_id);
CREATE INDEX idx_payment_status ON contribution_payment(status);
CREATE INDEX idx_payment_date ON contribution_payment(payment_date);
```

### Why?

* Fast lookup of payments per due
* Filtering successful vs failed
* Time-based analytics

---

## 🔄 Relationship

```text id="k92lwx"
ContributionDue (1) ---> ContributionPayment (many)
```

👉 One due can have:

* Multiple partial payments
* Multiple failed attempts
* Multiple retries

---

## 🧮 Derived Calculations (CRITICAL)

Never store totals inside this table.

### Total Paid:

```sql id="b3n7v1"
SELECT COALESCE(SUM(amount_paid), 0)
FROM contribution_payment
WHERE due_id = :dueId
AND status = 'SUCCESS';
```

---

### Remaining Amount:

```sql id="w6k3dp"
remaining = due.amount - total_paid
```

---

## 🔐 Transaction Handling (VERY IMPORTANT)

All payment operations must be transactional:

```sql id="p4s8yz"
BEGIN;

INSERT INTO contribution_payment (...);

-- Recalculate total paid
-- Update contribution_due.status

COMMIT;
```

👉 Prevents inconsistent states

---

## 🧪 Sample Inserts

### Successful Payment

```sql id="c7l9qa"
INSERT INTO contribution_payment 
(due_id, amount_paid, payment_mode, transaction_ref, status)
VALUES 
(1, 5000, 'UPI', 'UPI123ABC', 'SUCCESS');
```

---

### Failed Payment

```sql id="r5t2mz"
INSERT INTO contribution_payment 
(due_id, amount_paid, payment_mode, status)
VALUES 
(1, 5000, 'UPI', 'FAILED');
```

---

## 🚨 Important Notes

* Never delete payment records
* Never update historical transactions
* Always log failed attempts
* Always use backend/service layer for inserts

---

## 🔮 Future Extensibility

You can later add:

* `gateway_response` (JSON)
* `refund_amount`
* `is_refunded`
* `payment_gateway` (Razorpay, Stripe, etc.)
* `metadata` (for extensibility)

---

## 🧩 Integration with ContributionDue

After every successful payment:

### Logic:

```text id="7l2mnd"
if total_paid >= due.amount:
    due.status = PAID
else if today > due_date:
    due.status = OVERDUE
else:
    due.status = PENDING
```

---

## ✅ Summary

| Aspect        | Decision                  |
| ------------- | ------------------------- |
| Table Type    | Transaction               |
| Nature        | Append-only               |
| Cardinality   | Many per due              |
| Financial Use | Only SUCCESS records      |
| Dependency    | Linked to ContributionDue |

---