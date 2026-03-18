# User Management - Business Logic Layer (Elastic Band Up)

## Responsibility

Handles user profile management, RBAC, admin operations. Sits between data access and application (API) layer.

## Files (Production)

- `backend/user-management/handlers/admin.go` - Admin user management, RBAC
- `backend/internal/middleware/middleware.go` - Role checking middleware
- `backend/internal/services/` - User-related services
- `backend/internal/routes/` - User profile, admin routes

## Key Operations

- Profile CRUD validation
- RBAC permission checks
- Admin user listing, suspend/activate
- Role assignment logic

## Dependencies

- Data Access: User models, profile models
- Authentication: JWT validation, user identity

## Used By

- Application layer: API routes at `/api/v1/users/*`, `/api/v1/admin/users/*`

## Notes

Admin handler is large (~3000+ lines). Consider splitting by admin subsystem in future refactor.
