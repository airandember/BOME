# User Management - Data Access Layer (Elastic Band Up)

## Responsibility

Maps database operations for user profiles, preferences, and admin operations to Go models. Sits between persistence (users, user_preferences tables) and business logic.

## Files (Production)

- `backend/user-management/models/user-profile.go` - UserProfile model, preferences
- `backend/internal/database/user.go` - User CRUD (shared with auth)
- `backend/internal/database/` - User-related queries

## Key Operations

- Get/Update user profile
- Get/Update user preferences
- Admin: List users, get user by ID
- Role and permission lookups

## Dependencies

- Persistence: users, user_preferences, user_roles tables
- Authentication braid: User model, session validation

## Used By

- Business logic: `backend/user-management/handlers/admin.go`
- Routes: `backend/internal/routes` (user profile, admin endpoints)

## Notes

User Management extends the Authentication braid's users table. Profile and preference data may live in users table or separate user_preferences table.
