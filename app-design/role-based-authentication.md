# 🔐 Role-Based Authorization Design — Kitty Party App

## 🎯 Objective

Implement a **role-based authorization system** where:

* A **Global Admin (Super Admin)** has access to all data and actions
* Regular users have **context-based roles** within groups
* Permissions are enforced both at **API level** and **UI level**

---

## 🧠 Core Concepts

### 1. Global Role vs Group Role

We separate authorization into two layers:

#### 🌍 Global Roles

| Role       | Description                                    |
| ---------- | ---------------------------------------------- |
| SuperAdmin | Full access to all groups, users, transactions |
| User       | Default role with limited access               |

#### 👥 Group Roles (Contextual Roles)

| Role      | Description                                            |
| --------- | ------------------------------------------------------ |
| Admin     | Manages group (add/remove members, schedule, settings) |
| Treasurer | Handles money distribution, contributions              |
| Member    | Regular participant                                    |
| Viewer    | Read-only access                                       |

---

## 🧱 Database Design

### 🧾 Users Table

```json
{
  "id": "uuid",
  "name": "Jay",
  "email": "jay@example.com",
  "global_role": "SuperAdmin" // or "User"
}
```

---

### 👥 Groups Table

```json
{
  "id": "uuid",
  "name": "Office Kitty",
  "created_by": "user_id"
}
```

---

### 🔗 GroupMembers Table (IMPORTANT)

This is where contextual roles live.

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "group_id": "uuid",
  "role": "Admin", // Admin | Treasurer | Member | Viewer
  "joined_at": "datetime"
}
```

👉 Same user can have **different roles in different groups**

---

## 🔑 Authorization Strategy

### 1. JWT Token Structure

```json
{
  "user_id": "uuid",
  "global_role": "SuperAdmin"
}
```

👉 Keep JWT lightweight (DO NOT store group roles here)

---

### 2. Authorization Flow

#### Step 1: Authenticate User

* Validate JWT
* Extract `user_id` and `global_role`

#### Step 2: Check Global Role

```csharp
if (user.GlobalRole == "SuperAdmin")
{
    return AllowAllAccess();
}
```

---

#### Step 3: Fetch Group Role (if needed)

```sql
SELECT role 
FROM GroupMembers 
WHERE user_id = @userId AND group_id = @groupId
```

---

#### Step 4: Apply Permission Rules

Example:

```csharp
if (groupRole == "Admin")
{
    allow: AddMember, RemoveMember, UpdateGroup
}

if (groupRole == "Treasurer")
{
    allow: ManagePayments, DistributeMoney
}

if (groupRole == "Member")
{
    allow: ViewGroup, Participate
}

if (groupRole == "Viewer")
{
    allow: ViewOnly
}
```

---

## 🧩 Backend Implementation (C# .NET)

### 1. Create Authorization Attribute

```csharp
public class GroupRoleAuthorizeAttribute : AuthorizeAttribute
{
    public string[] AllowedRoles { get; set; }
}
```

---

### 2. Custom Authorization Handler

```csharp
public class GroupRoleHandler : AuthorizationHandler<GroupRoleRequirement>
{
    protected override async Task HandleRequirementAsync(
        AuthorizationHandlerContext context,
        GroupRoleRequirement requirement)
    {
        var userId = context.User.FindFirst("user_id")?.Value;
        var groupId = GetGroupIdFromRoute(context);

        var role = await _repo.GetUserRoleInGroup(userId, groupId);

        if (requirement.AllowedRoles.Contains(role))
        {
            context.Succeed(requirement);
        }
    }
}
```

---

### 3. Usage in Controller

```csharp
[GroupRoleAuthorize(AllowedRoles = new[] { "Admin" })]
[HttpPost("add-member")]
public IActionResult AddMember()
{
    return Ok();
}
```

---

## 🖥️ Frontend Authorization (React)

### 1. Store User Info

```js
{
  userId: "uuid",
  globalRole: "User"
}
```

---

### 2. Fetch Group Role

```js
GET /groups/{groupId}/my-role
```

---

### 3. Conditional UI Rendering

```js
if (groupRole === "Admin") {
  showAddMemberButton = true;
}
```

---

## 🚀 Advanced Ideas (Highly Recommended)

### 🔥 1. Permission-Based System (Future Upgrade)

Instead of roles → use permissions:

```json
{
  "can_add_member": true,
  "can_manage_money": false
}
```

---

### 🔥 2. Audit Logs

Track who did what:

```json
{
  "action": "ADD_MEMBER",
  "performed_by": "user_id",
  "group_id": "uuid",
  "timestamp": "datetime"
}
```

---

### 🔥 3. Role Change History

Track trust issues:

```json
{
  "old_role": "Member",
  "new_role": "Treasurer",
  "changed_by": "Admin"
}
```

---

### 🔥 4. Temporary Roles

Example:

* "Treasurer for this month only"

---

## ⚠️ Important Rules

* NEVER trust frontend role checks → always validate in backend
* SuperAdmin should bypass all checks
* Always validate `group_id` from request
* Avoid storing group roles in JWT (they change frequently)

---

## 🧪 Example Scenarios

### ✅ Scenario 1

User is **SuperAdmin**
→ Access everything

---

### ✅ Scenario 2

User is **Admin of Group A**
→ Can add/remove members in Group A only

---

### ❌ Scenario 3

User is **Member of Group B**
→ Cannot add members

---

## 🧭 Summary

* Use **Global Role + Group Role**
* Store group roles in **GroupMembers table**
* Keep JWT minimal
* Enforce authorization in backend using custom handlers
* Control UI using role-based rendering

---

**Done ✅ — You can directly implement this**
