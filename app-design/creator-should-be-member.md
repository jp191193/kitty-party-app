# 📘 Auto Member Assignment for Group Creator (Kitty App)

## 🎯 Purpose

Ensure that when a group is created, the **creator is automatically added as a member of that group**, without requiring a separate API call.

This eliminates:

* Redundant operations
* Data inconsistency
* Missing ownership scenarios

---

## 🧠 Core Principle

> A group cannot exist without its creator being a member of it.

The creator is:

* The **first member**
* The **admin by default**
* The **owner of the group**

---

## 🧱 Data Assumption

### Group Table

```json
{
  "id": 101,
  "created_by": "a6c23e3e-0b4b-48dc-85b3-f9b4d43b347e"
}
```

---

### GroupMember Table

```json
{
  "group_id": 101,
  "member_id": "a6c23e3e-0b4b-48dc-85b3-f9b4d43b347e",
  "role": "ADMIN",
  "status": "ACTIVE"
}
```

---

## ⚙️ Functional Requirement

### When:

* A new group is created

### Then:

* Automatically insert creator into `group_members`

---

## 🔄 Flow

```text
1. Create Group
2. Insert GroupMember (creator)
3. Commit transaction
```

---

## ⚡ Implementation Rule (CRITICAL)

This must happen:

* Inside the **same transaction**
* In the **same service layer method**

---

```csharp
using (var transaction = dbContext.Database.BeginTransaction())
{
    var group = new Group
    {
        Name = request.Name,
        CreatedBy = request.UserId
    };

    dbContext.Groups.Add(group);
    dbContext.SaveChanges();

    var groupMember = new GroupMember
    {
        GroupId = group.Id,
        MemberId = request.UserId,
        Role = "ADMIN",
        Status = "ACTIVE",
        JoinedAt = DateTime.UtcNow
    };

    dbContext.GroupMembers.Add(groupMember);
    dbContext.SaveChanges();

    transaction.Commit();
}
```

---

## 🔒 Constraints (MANDATORY)

### 1. Prevent Duplicate Membership

```sql
UNIQUE (group_id, member_id)
```

✅ Ensures creator is not added twice

---

### 2. Role Enforcement

* Creator must always have:

  * `role = ADMIN`
  * Cannot be downgraded initially (optional rule)

---

## 📊 Status Definition

| Status  | Meaning                   |
| ------- | ------------------------- |
| ACTIVE  | Member is active in group |
| INVITED | Awaiting acceptance       |
| LEFT    | Member exited group       |

---

## 🚨 Important Rules

### Rule 1: No Separate API Required

❌ Do NOT call "Add Member" API for creator
✅ Must be automatic

---

### Rule 2: Atomic Operation

* If member creation fails → group creation must rollback

---

### Rule 3: Creator Cannot Be Missing

* Every group must always have at least 1 member (creator)

---

### Rule 4: Creator Cannot Be Invited

* Creator is directly `ACTIVE`, not `INVITED`

---

## 🧩 Edge Cases

### Case 1: Retry Group Creation

* Ensure idempotency or transaction rollback

---

### Case 2: Duplicate Insert Attempt

* Handled via unique constraint

---

### Case 3: Creator Deletion (Future Rule)

* If creator leaves:

  * Transfer admin role
  * OR restrict leaving

---

## 🔮 Future Enhancements

* Multiple admins
* Role hierarchy (ADMIN / TREASURER / MEMBER)
* Ownership transfer
* Audit logs

---

## ✅ Summary

| Aspect          | Decision                    |
| --------------- | --------------------------- |
| Creator Role    | ADMIN                       |
| Membership Type | Auto-created                |
| API Requirement | Not needed                  |
| Transaction     | Mandatory                   |
| Constraint      | UNIQUE(group_id, member_id) |

---