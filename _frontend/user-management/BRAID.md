# 🧬 User Management Braid - Frontend
**Svelte5 UI for user profiles, preferences, and RBAC**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **frontend portion** of the User Management Braid.  
> **Backend portion**: See `_braids/user-management/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Frontend Overview**

**Purpose**: User interface for profile management, settings, and admin operations  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/account/profile`, `/account/settings`, `/admin/users`  
**State Management**: Svelte stores with reactive profile state

---

## 🎯 **Key Features**

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

## 📄 **Frontend Pages**

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

## 🧩 **Frontend Components**

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

## 🗃️ **Frontend Stores**

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

## 🎨 **UI Patterns**

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

## 🔄 **Data Flow Examples**

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

## 🔒 **Security Considerations**

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

## 📊 **Performance Optimizations**

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

## 🎯 **Accessibility**

### **Features**:
- Keyboard navigation
- Screen reader labels
- ARIA attributes
- Focus management
- Color contrast compliance
- Form validation announcements

---

## 🔗 **Related Documentation**

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

## 📝 **Known Issues**

### **To Address**:
1. Avatar upload could use better cropping UI
2. Profile completion progress needs implementation
3. Advanced search filters limited
4. Bulk operations UI could be improved
5. Activity feed performance on large datasets

---

## 🚀 **Quick Links**

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
[🏠 Master Index](../../../BRAIDS_INDEX.md) | [⬅️ Backend Braid](../../_braids/user-management/backend/BRAID.md) | [🔗 Auth Braid](../authentication/BRAID.md)

