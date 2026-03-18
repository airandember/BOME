# Braid: user-management

**Architecture:** Full-Stack Braid (Frontend to Backend)
**Last Updated:** 2025-10-17

---

## Backend Architecture

**User profiles, preferences, RBAC, and account management**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **backend portion** of the User Management Braid.  
> **Frontend portion**: See `_frontend/braids/user-management/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Backend Overview**

**Purpose**: Server-side user profile management, RBAC, and account operations  
**Technology**: Go, PostgreSQL, JWT-based authorization  
**Complexity**: Medium-High (RBAC, Profile Management, Admin Operations)  
**Dependencies**: Authentication Braid (user identity)

---

## 📁 **Production File Map**

### **Backend Files (Go)**
```
backend/
├── user-management/
│   ├── handlers/admin.go              # Admin user management routes
│   └── models/user-profile.go         # User profile model
├── internal/
│   ├── routes/                         # User profile, admin routes
│   ├── database/user.go               # User DB operations
│   └── middleware/                     # RBAC middleware
└── authentication/
    └── models/user.go                  # Shared user model (from auth braid)
```

### **Frontend Files (Svelte)**
```
frontend/src/
├── routes/
│   ├── account/                        # User account pages
│   ├── profile/                        # Profile pages
│   └── settings/                       # Settings pages
└── lib/
    └── components/                     # Profile, settings components
```

---

## ðŸŽ¯ **Key Features**

### **1. User Profile Management**:
- Profile CRUD operations
- Avatar/image upload
- Bio and personal information
- Profile picture management
- Account settings

### **2. Role-Based Access Control (RBAC)**:
- User roles (user, admin, moderator, etc.)
- Permission checking
- Role assignments
- Hierarchical permissions
- Middleware enforcement

### **3. User Preferences**:
- Theme preferences (light/dark)
- Notification settings
- Privacy settings
- Language preferences
- Email preferences

### **4. Admin User Management**:
- User listing and search
- User account suspension/activation
- Role management
- User deletion (GDPR compliance)
- Bulk operations

### **5. Activity Tracking**:
- User action logging
- Audit trail
- Login history
- Activity dashboard
- Security monitoring

---

## ðŸ—„ï¸ **Database Schema**

### **Core Tables**:

**users** (Extended from auth braid):
- `id` - Primary key
- `email` - Email address (unique)
- `first_name` - First name
- `last_name` - Last name
- `profile_picture_url` - Avatar URL
- `bio` - User biography
- `role` - User role (user/admin/etc.)
- `is_active` - Account status
- `email_verified` - Email verification status
- `created_at` - Account creation
- `updated_at` - Last update
- `last_login` - Last login timestamp

**user_preferences**:
- `id` - Primary key
- `user_id` - Foreign key to users
- `theme` - UI theme (light/dark/auto)
- `language` - Preferred language
- `timezone` - User timezone
- `email_notifications` - Email notification settings
- `push_notifications` - Push notification settings
- `privacy_settings` - JSON privacy configuration
- `created_at` - Created timestamp
- `updated_at` - Updated timestamp

**user_roles** (RBAC):
- `id` - Primary key
- `name` - Role name (unique)
- `description` - Role description
- `permissions` - JSON array of permissions
- `is_admin` - Admin flag
- `created_at` - Created timestamp

**user_activity_log**:
- `id` - Primary key
- `user_id` - Foreign key to users
- `action` - Action performed
- `entity_type` - Entity affected
- `entity_id` - Entity ID
- `details` - JSON action details
- `ip_address` - User IP
- `user_agent` - Browser/client
- `created_at` - Action timestamp

---

## ðŸŒ **API Endpoints**

### **Profile Management**:
```
GET    /api/v1/users/profile        # Get current user profile
PUT    /api/v1/users/profile        # Update profile
POST   /api/v1/users/avatar         # Upload avatar
DELETE /api/v1/users/avatar         # Delete avatar
GET    /api/v1/users/:id/profile    # Get user by ID (public)
```

### **Preferences**:
```
GET    /api/v1/users/preferences    # Get user preferences
PUT    /api/v1/users/preferences    # Update preferences
PATCH  /api/v1/users/preferences/:key  # Update single preference
```

### **Admin User Management**:
```
GET    /api/v1/admin/users          # List users (paginated)
GET    /api/v1/admin/users/:id      # Get user details
PUT    /api/v1/admin/users/:id      # Update user
DELETE /api/v1/admin/users/:id      # Delete user (soft delete)
POST   /api/v1/admin/users/:id/suspend    # Suspend user
POST   /api/v1/admin/users/:id/activate   # Activate user
PUT    /api/v1/admin/users/:id/role       # Change user role
```

### **Roles & Permissions**:
```
GET    /api/v1/roles                # List all roles
GET    /api/v1/roles/:id            # Get role details
POST   /api/v1/roles                # Create role (admin)
PUT    /api/v1/roles/:id            # Update role (admin)
DELETE /api/v1/roles/:id            # Delete role (admin)
GET    /api/v1/users/:id/permissions  # Get user permissions
```

### **Activity Tracking**:
```
GET    /api/v1/users/activity       # Get user activity log
GET    /api/v1/admin/activity       # Get all activity (admin)
```

---

## ðŸ”— **Integration with Authentication Braid**

### **Shared Concepts**:
- **Users table**: Extended from auth braid
- **JWT tokens**: Include role and permissions
- **Auth middleware**: Enhanced with RBAC checks
- **Session tracking**: Linked to user profiles

### **Dependencies**:
```
Authentication Braid â†’ User Management Braid
â”œâ”€â”€ User identity (user_id)
â”œâ”€â”€ Email verification status
â”œâ”€â”€ JWT token claims (role, permissions)
â”œâ”€â”€ Session management
â””â”€â”€ Auth middleware
```

---

## ðŸ“ **File Structure**

### **Backend Routes** (`backend/internal/routes/`):
- `user.go` - User profile endpoints
- `admin.go` - Admin user management (user section)
- `roles.go` - RBAC role management

### **Backend Services** (`backend/internal/services/`):
- User profile business logic
- RBAC permission checking
- Activity logging

### **Backend Database** (`backend/internal/database/`):
- `user.go` - User operations
- Role and permission operations
- Activity log operations

### **Migrations** (`backend/migrations/`):
- `005_add_user_profile_fields.sql`
- User preferences table
- User roles table
- Activity log table

---

## ðŸ”’ **RBAC System**

### **Built-in Roles**:
1. **user** - Standard user (default)
2. **admin** - Full system access
3. **moderator** - Content moderation
4. **support** - Customer support access
5. **analyst** - Analytics and reporting

### **Permission Categories**:
- `users:read` - View user profiles
- `users:write` - Edit users
- `users:delete` - Delete users
- `content:moderate` - Moderate content
- `analytics:view` - View analytics
- `system:admin` - System administration

### **Permission Checking**:
```go
// In middleware
func RequirePermission(permission string) middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user").(*User)
            
            if !user.HasPermission(permission) {
                respondError(w, "Permission denied", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
router.Handle("/admin/users", RequirePermission("users:write"))
```

---

## ðŸŽ¨ **User Profile Fields**

### **Basic Information**:
- First name, last name
- Email (from auth)
- Profile picture URL
- Bio/description
- Location (optional)
- Website (optional)
- Social media links

### **Account Information**:
- Role and permissions
- Account creation date
- Last login timestamp
- Account status (active/suspended)
- Email verification status

### **Preferences**:
- UI theme
- Language
- Timezone
- Notification settings
- Privacy settings

---

## ðŸ”„ **Data Flow Examples**

### **Profile Update Flow**:
```
1. User edits profile in frontend
2. PUT /api/v1/users/profile
3. Validate input data
4. Check auth (JWT middleware)
5. Update users table
6. Log activity
7. Return updated profile
8. Frontend updates store
```

### **Admin User Suspension**:
```
1. Admin clicks "Suspend User"
2. POST /api/v1/admin/users/:id/suspend
3. Check admin permission
4. Update is_active = false
5. Invalidate user sessions
6. Log admin action
7. Send notification email
8. Return success
```

### **RBAC Permission Check**:
```
1. Request to protected endpoint
2. Auth middleware validates JWT
3. Extract user_id and role
4. RBAC middleware checks permission
5. Query user permissions from role
6. Allow or deny access
7. Log access attempt
8. Proceed or return 403
```

---

## ðŸ“Š **Performance Considerations**

### **Caching**:
- User profiles cached (5 min TTL)
- Role permissions cached (15 min TTL)
- Activity logs not cached

### **Database Indexes**:
- `users.email` (unique index)
- `users.role` (for role queries)
- `user_activity_log.user_id` (for activity queries)
- `user_activity_log.created_at` (for time-based queries)

### **Pagination**:
- User listing paginated (50 per page)
- Activity log paginated (100 per page)
- Search results paginated (20 per page)

---

## ðŸ” **Security Measures**

### **Data Protection**:
- Profile updates require authentication
- Admin operations require admin role
- Sensitive fields (email) require additional verification
- GDPR compliance (right to be forgotten)

### **Access Control**:
- RBAC enforced at middleware level
- Permission checks before sensitive operations
- Audit log for all admin actions
- Session invalidation on role change

### **Privacy**:
- Optional profile visibility settings
- Email privacy controls
- Activity log retention (90 days)
- User data export capability

---

## ðŸ§¬ **Strands (Complete Flows)**

### **1. Profile Management Strand**:
Complete flow from profile edit form to database update

### **2. RBAC System Strand**:
Permission checking from request to database

### **3. User Preferences Strand**:
Settings management from UI to storage

### **4. Admin User Management Strand**:
Admin operations from dashboard to database

---

## ðŸ“ **Known Technical Debt**

### **To Address**:
1. Profile pictures stored externally (Digital Ocean Spaces)
2. Activity log grows large (need archival strategy)
3. Role permissions could use more granularity
4. User search could be optimized with ElasticSearch
5. Profile validation could be more comprehensive

### **Future Enhancements**:
1. Two-factor authentication integration
2. Social login profile sync
3. Advanced privacy controls
4. User blocking/muting features
5. Profile verification badges

---

## ðŸ”— **Related Braids**

### **Depends On**:
- **Authentication Braid**: User identity, JWT tokens
- **Infrastructure Braid**: Database, file storage

### **Consumed By**:
- **Admin Dashboard Braid**: User management UI
- **Subscription Braid**: User billing info
- **Analytics Braid**: User behavior tracking
- **Content Management Braid**: User-generated content

---

## ðŸ“š **Documentation Files**

### **Layer Documentation**:
- `layers/persistence/schema/` - Database schema docs
- `layers/data-access/` - Go model operations
- `layers/business-logic/` - API handlers and services
- `layers/application/` - API contracts

### **Strand Documentation**:
- `strands/profile-management/` - Profile CRUD flow
- `strands/rbac-system/` - Permission checking flow
- `strands/user-preferences/` - Settings management
- `strands/admin-user-management/` - Admin operations

### **Elastic Bands**:
- `layers/*/ELASTIC-BAND-UP.md` - Upward contracts
- `layers/*/ELASTIC-BAND-DOWN.md` - Downward contracts

---

## ðŸš€ **Quick Start**

### **Understanding User Management**:
1. Read this BRAID.md (15 min)
2. Check `strands/profile-management/STRAND.md` (15 min)
3. Review RBAC system in `strands/rbac-system/STRAND.md` (15 min)
4. **Total: 45 minutes to understand user management!**

### **Working on Features**:
1. Identify which strand (profile, RBAC, preferences, admin)
2. Read relevant strand document
3. Check elastic bands for layer contracts
4. Follow existing patterns
5. Update documentation

---

**Last Updated**: October 14, 2025  
**Status**: Complete backend documentation  
**Technology**: Go + PostgreSQL  
**Frontend Counterpart**: `_frontend/braids/user-management/`

---

**Navigate**:  
[ðŸ  Master Index](../../BRAIDS_INDEX.md) | [ðŸŽ¨ Frontend Braid](../../_frontend/braids/user-management/BRAID.md) | [ðŸ”— Auth Braid](../authentication/BRAID.md)



---

## Frontend Architecture

**Svelte5 UI for user profiles, preferences, and RBAC**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **frontend portion** of the User Management Braid.  
> **Backend portion**: See `_braids/user-management/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Frontend Overview**

**Purpose**: User interface for profile management, settings, and admin operations  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/account/profile`, `/account/settings`, `/admin/users`  
**State Management**: Svelte stores with reactive profile state

---

## ðŸŽ¯ **Key Features**

### **1. User Profile Pages**:
- Profile viewing and editing
- Avatar upload and management
- Bio and personal info
- Public profile view
- Profile completion progress

### **2. Account Settings**:
- Theme preferences (light/dark/auto)
- Language selection
- Timezone settings
- Email notification preferences
- Privacy controls

### **3. Role-Based UI**:
- Conditional rendering based on permissions
- Admin-only menu items
- Role badges and indicators
- Permission-aware components

### **4. Admin User Management**:
- User listing with search and filters
- User detail view
- Suspend/activate users
- Role management interface
- Bulk operations

---

## ðŸ“„ **Frontend Pages**

### **1. Profile Page** (`/account/profile`)
**File**: `frontend/src/routes/account/profile/+page.svelte`

**Features**:
- View current profile
- Edit profile form
- Avatar upload with preview
- Bio editor with character count
- Save/cancel buttons
- Loading states
- Success/error messages

**API Calls**:
- `GET /api/v1/users/profile`
- `PUT /api/v1/users/profile`
- `POST /api/v1/users/avatar`

---

### **2. Settings Page** (`/account/settings`)
**File**: `frontend/src/routes/account/settings/+page.svelte`

**Features**:
- Theme selector (light/dark/auto)
- Language dropdown
- Timezone picker
- Notification toggles
- Privacy controls
- Save preferences button

**API Calls**:
- `GET /api/v1/users/preferences`
- `PUT /api/v1/users/preferences`

---

### **3. Public Profile View** (`/users/:id`)
**File**: `frontend/src/routes/users/[id]/+page.svelte`

**Features**:
- Display user public profile
- Avatar and bio
- Role badge
- Activity feed (if public)
- Social links

**API Calls**:
- `GET /api/v1/users/:id/profile`

---

### **4. Admin Users Page** (`/admin/users`)
**File**: `frontend/src/routes/admin/users/+page.svelte`

**Features**:
- User listing table
- Search and filters
- Pagination
- Sort options
- User status badges
- Quick actions (suspend, activate)
- Bulk selection

**API Calls**:
- `GET /api/v1/admin/users`
- `POST /api/v1/admin/users/:id/suspend`
- `POST /api/v1/admin/users/:id/activate`

---

### **5. User Detail Page** (`/admin/users/:id`)
**File**: `frontend/src/routes/admin/users/[id]/+page.svelte`

**Features**:
- Complete user information
- Edit user details
- Change role
- View activity log
- Delete user
- Send notifications

**API Calls**:
- `GET /api/v1/admin/users/:id`
- `PUT /api/v1/admin/users/:id`
- `PUT /api/v1/admin/users/:id/role`
- `DELETE /api/v1/admin/users/:id`

---

## ðŸ§© **Frontend Components**

### **ProfileForm Component**
**Purpose**: Reusable profile editing form

**Props**:
```typescript
interface ProfileFormProps {
  initialData: UserProfile;
  onSave: (data: UserProfile) => Promise<void>;
  onCancel: () => void;
  readonly?: boolean;
}
```

**Features**:
- Form validation
- Character limits
- Real-time validation
- Loading states
- Error display

---

### **AvatarUpload Component**
**Purpose**: Avatar image upload and management

**Features**:
- Drag and drop
- File selection
- Image preview
- Crop/resize (future)
- Delete avatar
- Loading indicator

---

### **RoleBadge Component**
**Purpose**: Display user role with styling

**Props**:
```typescript
interface RoleBadgeProps {
  role: string;
  size?: 'small' | 'medium' | 'large';
}
```

**Styling**:
- Admin: Red badge
- Moderator: Orange badge
- Support: Blue badge
- User: Gray badge

---

### **PermissionGate Component**
**Purpose**: Conditional rendering based on permissions

**Usage**:
```svelte
<PermissionGate permission="users:write">
  <button>Edit User</button>
</PermissionGate>
```

---

### **UserTable Component**
**Purpose**: Reusable user listing table

**Features**:
- Sortable columns
- Pagination
- Search integration
- Bulk selection
- Action buttons
- Status indicators

---

## ðŸ—ƒï¸ **Frontend Stores**

### **User Profile Store** (`$lib/stores/userProfile.ts`)
**Purpose**: Manage current user profile state

**State**:
```typescript
interface UserProfileState {
  profile: UserProfile | null;
  loading: boolean;
  error: string | null;
  isEditing: boolean;
}
```

**Methods**:
- `loadProfile()` - Fetch current user profile
- `updateProfile(data)` - Update profile
- `uploadAvatar(file)` - Upload avatar image
- `deleteAvatar()` - Remove avatar

---

### **User Preferences Store** (`$lib/stores/userPreferences.ts`)
**Purpose**: Manage user preferences and settings

**State**:
```typescript
interface UserPreferencesState {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  timezone: string;
  notifications: NotificationSettings;
  privacy: PrivacySettings;
  loading: boolean;
}
```

**Methods**:
- `loadPreferences()` - Fetch preferences
- `updatePreferences(data)` - Update all preferences
- `updateTheme(theme)` - Update theme only
- `applyTheme()` - Apply theme to UI

---

### **Admin Users Store** (`$lib/stores/adminUsers.ts`)
**Purpose**: Manage admin user listing and operations

**State**:
```typescript
interface AdminUsersState {
  users: User[];
  total: number;
  page: number;
  pageSize: number;
  search: string;
  filters: UserFilters;
  loading: boolean;
  selected: Set<string>;
}
```

**Methods**:
- `loadUsers()` - Fetch user list
- `searchUsers(query)` - Search users
- `suspendUser(id)` - Suspend user
- `activateUser(id)` - Activate user
- `deleteUser(id)` - Delete user
- `bulkAction(action)` - Perform bulk operation

---

## ðŸŽ¨ **UI Patterns**

### **Profile Editing**:
```svelte
<script>
  import { userProfile } from '$lib/stores/userProfile';
  
  let editing = false;
  let formData = { ...$userProfile.profile };
  
  async function handleSave() {
    try {
      await userProfile.updateProfile(formData);
      editing = false;
      showSuccess('Profile updated!');
    } catch (error) {
      showError(error.message);
    }
  }
</script>

{#if editing}
  <ProfileForm
    initialData={formData}
    bind:data={formData}
    onSave={handleSave}
    onCancel={() => editing = false}
  />
{:else}
  <ProfileView data={$userProfile.profile} />
  <button on:click={() => editing = true}>Edit Profile</button>
{/if}
```

---

### **Permission-Based Rendering**:
```svelte
<script>
  import { auth } from '$lib/auth';
  
  $: hasAdminAccess = $auth.user?.role === 'admin';
  $: canEditUsers = hasPermission('users:write');
</script>

{#if hasAdminAccess}
  <a href="/admin/users">Admin Panel</a>
{/if}

{#if canEditUsers}
  <button on:click={editUser}>Edit</button>
{/if}
```

---

### **Theme Management**:
```svelte
<script>
  import { userPreferences } from '$lib/stores/userPreferences';
  
  $: theme = $userPreferences.theme;
  
  // Apply theme to document
  $: {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else if (theme === 'light') {
      document.documentElement.classList.remove('dark');
    } else {
      // Auto: follow system preference
      const darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
      document.documentElement.classList.toggle('dark', darkMode);
    }
  }
  
  function setTheme(newTheme: string) {
    userPreferences.updateTheme(newTheme);
  }
</script>

<select value={theme} on:change={(e) => setTheme(e.target.value)}>
  <option value="light">Light</option>
  <option value="dark">Dark</option>
  <option value="auto">Auto</option>
</select>
```

---

## ðŸ”„ **Data Flow Examples**

### **Profile Update Flow**:
```
1. User clicks "Edit Profile"
2. Component enters edit mode
3. User modifies fields
4. User clicks "Save"
5. Store.updateProfile() called
6. API: PUT /api/v1/users/profile
7. Backend validates & updates
8. Success response
9. Store updates reactive state
10. UI reflects changes
11. Success message shown
```

### **Admin User Suspension**:
```
1. Admin clicks "Suspend" on user
2. Confirmation dialog shown
3. Admin confirms
4. Store.suspendUser(id) called
5. API: POST /api/v1/admin/users/:id/suspend
6. Backend updates is_active = false
7. Success response
8. User list refreshed
9. User status badge updated
10. Success notification
```

### **Theme Change**:
```
1. User selects theme from dropdown
2. Store.updateTheme() called
3. API: PATCH /api/v1/users/preferences/theme
4. Backend saves preference
5. Store updates reactive state
6. $: reactive block triggers
7. document.documentElement.classList updated
8. Theme applied immediately
9. Success saved to backend
```

---

## ðŸ”’ **Security Considerations**

### **Client-Side Validation**:
- Email format validation
- Bio character limits (500 chars)
- Username format validation
- File type validation (images only)
- File size limits (5MB max)

### **Permission Checks**:
- UI elements hidden without permissions
- API calls require authentication
- Role-based route guards
- Admin pages require admin role

### **Data Privacy**:
- Sensitive data not cached
- Profile visibility controls
- Optional field display
- User consent for data usage

---

## ðŸ“Š **Performance Optimizations**

### **Lazy Loading**:
- Admin pages code-split
- Profile images lazy loaded
- User avatars with loading placeholders

### **State Management**:
- Profile cached (5 min)
- Preferences cached (15 min)
- Reactive updates minimize re-renders

### **API Efficiency**:
- Debounced search inputs
- Paginated user lists
- Optimistic UI updates

---

## ðŸŽ¯ **Accessibility**

### **Features**:
- Keyboard navigation
- Screen reader labels
- ARIA attributes
- Focus management
- Color contrast compliance
- Form validation announcements

---

## ðŸ”— **Related Documentation**

### **Backend Portion**:
- [`_braids/user-management/backend/BRAID.md`](../../_braids/user-management/backend/BRAID.md)

### **API Contracts**:
- [`_backend/.../layers/application/ELASTIC-BAND-UP.md`](../../_braids/user-management/backend/layers/application/ELASTIC-BAND-UP.md)

### **Strands**:
- Profile Management flow
- RBAC system flow
- Preferences management
- Admin operations

---

## ðŸ“ **Known Issues**

### **To Address**:
1. Avatar upload could use better cropping UI
2. Profile completion progress needs implementation
3. Advanced search filters limited
4. Bulk operations UI could be improved
5. Activity feed performance on large datasets

---

## ðŸš€ **Quick Links**

**Actual Files**:
- Profile Page: `frontend/src/routes/account/profile/+page.svelte`
- Settings Page: `frontend/src/routes/account/settings/+page.svelte`
- Admin Users: `frontend/src/routes/admin/users/+page.svelte`
- Profile Store: `frontend/src/lib/stores/userProfile.ts`
- Preferences Store: `frontend/src/lib/stores/userPreferences.ts`

---

**Last Updated**: October 14, 2025  
**Status**: Complete frontend documentation  
**Technology**: Svelte 5 + TypeScript  
**Backend Counterpart**: `_braids/user-management/backend/`

---

**Navigate**:  
[ðŸ  Master Index](../../../BRAIDS_INDEX.md) | [â¬…ï¸ Backend Braid](../../_braids/user-management/backend/BRAID.md) | [ðŸ”— Auth Braid](../authentication/BRAID.md)



---

## Integration Notes

- Frontend: `_braids/user-management/frontend/`
- Backend: `_braids/user-management/backend/`

This braid represents a complete vertical slice of functionality.

