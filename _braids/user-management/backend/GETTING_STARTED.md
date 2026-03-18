# Getting Started with User Management Braid

**Quick guide to user profiles, RBAC, and admin operations**

---

## What is This?

The User Management Braid handles user profiles, role-based access control (RBAC), preferences, and admin user operations. It extends the Authentication Braid with profile data and permissions.

---

## Quick Start

### Key Production Files

- **Handlers**: `backend/user-management/handlers/admin.go`
- **Models**: `backend/user-management/models/user-profile.go`
- **Routes**: `backend/internal/routes/` (admin, profile routes)
- **Database**: Users table (auth braid), user_preferences, user_roles

### Common Scenarios

**"User can't update profile"**
1. Check `backend/user-management/handlers/` for profile handlers
2. Verify JWT middleware allows authenticated users
3. Check `backend/internal/routes/` for profile route registration

**"Need to add new role"**
1. Review RBAC in `backend/authentication/middleware/middleware.go`
2. Check user_roles table schema in migrations
3. Add role to role assignment logic

**"Admin operations failing"**
1. Check `backend/user-management/handlers/admin.go`
2. Verify admin middleware/role check
3. Review `backend/internal/routes/admin.go`

---

## Dependencies

- **Authentication Braid**: User identity (JWT, user_id)
- **Database**: PostgreSQL with users, user_preferences, user_roles tables

---

## Documentation

- **BRAID.md**: Full braid overview at `_braids/user-management/BRAID.md`
- **File Map**: Production file mapping in BRAID.md

---

**Last Updated**: 2025-03-17
