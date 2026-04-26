# 📘 Authorization: Only Group Creator Can Add Members (JWT-Based)

## 🎯 Purpose

Ensure that **only the group creator (admin)** can add new members to a group.

This prevents:

* Unauthorized access
* Random users joining groups
* Security loopholes in group management

---

## 🧠 Core Principle

> Authorization must be enforced at the backend, not the frontend.

Even if frontend hides buttons, backend must validate:

* Who is making the request
* Whether they have permission

---

## 🔐 Authentication Mechanism

Use **JWT (JSON Web Token)** for identifying the user making the request.

### Token Payload Example

```json id="jwtpl1"
{
  "user_id": "c1e663a4-50df-435b-b8c4-b5eea7369ebf",
  "email": "jay@example.com",
  "exp": 1712345678
}
```

---

## ⚙️ Functional Requirement

### Scenario

User: **Jay Ponkia (c1e663a4-50df-435b-b8c4-b5eea7369ebf)**
Action: Add member to group

### Rule

✅ Allowed IF:

* User is the **creator of the group**
  OR
* User has **ADMIN role in group_members**

❌ Denied otherwise

---

## 🧱 Data Dependency

### Group Table

```json id="grp1"
{
  "id": 101,
  "created_by": "c1e663a4-50df-435b-b8c4-b5eea7369ebf"
}
```

---

### GroupMember Table

```json id="grp2"
{
  "group_id": 101,
  "member_id": "c1e663a4-50df-435b-b8c4-b5eea7369ebf",
  "role": "ADMIN"
}
```

---

## 🔄 Request Flow

```text id="flow1"
1. Client sends request with JWT token
2. Backend extracts user_id from token
3. Fetch group details
4. Validate:
    - Is user creator?
    OR
    - Is user ADMIN in group?
5. If valid → proceed
6. Else → return 403 Forbidden
```

---

## 📡 API Definition

```http id="api1"
POST /groups/{groupId}/members
Authorization: Bearer <JWT>
```

### Request Body

```json id="req1"
{
  "member_id": "new-user-id"
}
```

---

## 🧪 Validation Logic

### Step 1: Extract User from JWT

```csharp id="jwt1"
var userId = User.FindFirst("user_id")?.Value;
```

---

### Step 2: Fetch Group

```csharp id="grp3"
var group = dbContext.Groups
    .FirstOrDefault(g => g.Id == groupId);
```

---

### Step 3: Authorization Check

```csharp id="auth1"
bool isCreator = group.CreatedBy == userId;

bool isAdmin = dbContext.GroupMembers.Any(gm =>
    gm.GroupId == groupId &&
    gm.MemberId == userId &&
    gm.Role == "ADMIN"
);

if (!isCreator && !isAdmin)
{
    throw new UnauthorizedAccessException("Only admin can add members");
}
```

---

## 🔒 Security Rules (MANDATORY)

### Rule 1: Never Trust Frontend

❌ Do NOT rely on UI restrictions
✅ Always validate on backend

---

### Rule 2: JWT Must Be Validated

* Signature must be verified
* Token must not be expired

---

### Rule 3: Role-Based Authorization

* Do not rely only on `created_by`
* Always check `group_members.role`

---

### Rule 4: Return Proper HTTP Status

| Scenario     | Status Code |
| ------------ | ----------- |
| Unauthorized | 401         |
| Forbidden    | 403         |
| Success      | 200 / 201   |

---

## 🚨 Edge Cases

### Case 1: User Not in Group

❌ Cannot add members

---

### Case 2: Group Not Found

→ Return `404 Not Found`

---

### Case 3: Duplicate Member

→ Prevent using:

```sql id="dup1"
UNIQUE (group_id, member_id)
```

---

### Case 4: Self-Add

→ Optional: ignore or block

---

## 🧩 Optional Enhancement

### Use Policy-Based Authorization (.NET)

```csharp id="pol1"
services.AddAuthorization(options =>
{
    options.AddPolicy("GroupAdminOnly", policy =>
        policy.RequireAssertion(context =>
        {
            // custom logic here
            return true;
        }));
});
```

---

## 🔮 Future Enhancements

* Role hierarchy:

  * ADMIN
  * TREASURER
  * MEMBER

* Permission matrix:

  * Add member
  * Remove member
  * Declare winner

* Invite-based flow instead of direct add

---

## ✅ Summary

| Aspect             | Decision        |
| ------------------ | --------------- |
| Auth Mechanism     | JWT             |
| Authorization Type | Role-based      |
| Allowed Users      | Creator / ADMIN |
| Validation Layer   | Backend         |
| Failure Response   | 403 Forbidden   |

---