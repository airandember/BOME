# STRAND: User Management (Admin)

**Purpose**: Admin interface for viewing, editing, and managing user accounts.

---

## Implementation Details

### Backend
- **Handlers**: `backend/user-management/handlers/admin.go`
- **Routes**: `backend/internal/routes/admin.go` (admin user endpoints)
- **Services**: Various in `backend/internal/services/`
- **Database**: `users`, `user_roles`, `user_activity_log`

### Frontend
- **Pages**: `frontend/src/routes/admin/`

### Flow
1. Admin navigates to user management
2. Backend returns paginated user list (with RBAC check)
3. Admin can view details, suspend, change role
4. Actions logged to audit trail

---

## Status
- [x] Backend admin handlers implemented
- [x] User listing and update endpoints
- [ ] Bulk operations documented
- [ ] Activity log strand

---

## Testing
- Login as admin
- List users via API
- Suspend/activate user
- Verify RBAC enforcement
