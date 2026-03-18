# 🧬 Admin Dashboard Braid - Frontend
**Comprehensive admin interface for platform management**

---

## 📋 **Frontend Overview**

**Purpose**: Admin UI for managing users, content, subscriptions, and system operations  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/admin/*` (15+ pages)  

---

## 🎯 **Admin Pages** (15+ Areas)

### **1. Admin Dashboard** (`/admin`)
**File**: `frontend/src/routes/admin/+page.svelte`
- Key metrics cards
- Quick actions
- Recent activity
- System status

### **2. User Management** (`/admin/users`)
**File**: `frontend/src/routes/admin/users/+page.svelte`
- User list with filters
- Edit user dialog
- Assign roles
- Suspend/activate

### **3. Video Management** (`/admin/videos`)
**File**: `frontend/src/routes/admin/videos/+page.svelte`
- Video approval queue
- Edit video metadata
- Bulk operations

### **4. Analytics** (`/admin/analytics`)
**File**: `frontend/src/routes/admin/analytics/+page.svelte`
- Charts and graphs
- KPI cards
- Export reports

### **5. Subscriptions** (`/admin/streaming/subscriptions`)
**File**: `frontend/src/routes/admin/streaming/subscriptions/+page.svelte`
- Subscription list
- Revenue tracking
- Refund processing

### **6. Advertisements** (`/admin/advertisements`)
**File**: `frontend/src/routes/admin/advertisements/+page.svelte`
- Ad approval queue
- Performance metrics

### **7. Roles & Permissions** (`/admin/roles`)
- Role management
- Permission assignment
- RBAC configuration

### **8. System Monitoring** (`/admin/monitoring`)
- Health dashboards
- Error logs
- Performance metrics

### **9. Database Admin** (`/admin/database`)
- Database statistics
- Query performance
- Connection pools

### **10. Streaming Admin** (`/admin/streaming`)
- Video management
- CDN configuration
- Analytics

---

## 🧩 **Key Components**

### **AdminSidebar**:
```svelte
<nav class="admin-sidebar">
  <a href="/admin">Dashboard</a>
  <a href="/admin/users">Users</a>
  <a href="/admin/videos">Videos</a>
  <a href="/admin/analytics">Analytics</a>
  <a href="/admin/subscriptions">Subscriptions</a>
  <!-- ... more links -->
</nav>
```

### **DataTable** (Reusable):
**File**: `frontend/src/lib/components/DataTable.svelte`
- Sortable columns
- Pagination
- Filters
- Bulk actions
- Export to CSV

---

## 📊 **Admin Features**

### **Quick Actions**:
- Suspend user
- Approve content
- Issue refund
- Reset password
- Send notification

### **Bulk Operations**:
- Select multiple items
- Apply action to all selected
- Confirm dialog
- Progress tracking

---

## 🔒 **Role-Based UI**

### **Conditional Rendering**:
```svelte
{#if $auth.user?.hasPermission('edit_users')}
  <button on:click={editUser}>Edit</button>
{/if}

{#if $auth.user?.hasPermission('delete_users')}
  <button on:click={deleteUser}>Delete</button>
{/if}
```

---

## 📝 **Key Files**

**Admin Pages**:
- `frontend/src/routes/admin/+page.svelte` - Main dashboard
- `frontend/src/routes/admin/users/+page.svelte` - User management
- `frontend/src/routes/admin/analytics/+page.svelte` - Analytics
- `frontend/src/routes/admin/videos/+page.svelte` - Video management

---

**Last Updated**: October 14, 2025  
**Backend**: `_braids/admin-dashboard/backend/`

