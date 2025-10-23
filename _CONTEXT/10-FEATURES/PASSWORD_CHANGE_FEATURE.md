# 🔐 Forced Password Change Feature

## Overview

The BOME platform now includes a **forced password change** feature that requires users to change their default/temporary passwords on first login. This enhances security by ensuring that default credentials cannot be used indefinitely.

## 🚀 **How It Works**

### 1. **Database Schema**
- Added `password_changed` field to the `users` table
- Default value: `false` for new users
- Set to `true` after first password change

### 2. **User Creation**
- **Admin users**: Created with `password_changed = false`
- **Regular users**: Created with `password_changed = false`
- **Existing users**: Automatically set to `password_changed = true`

### 3. **Login Flow**
1. User logs in with credentials
2. System checks `password_changed` status
3. If `false`: Redirected to `/change-password`
4. If `true`: Normal login flow continues

### 4. **Password Change Process**
- User must provide current password
- New password must meet security requirements
- After successful change: `password_changed = true`
- User redirected to appropriate dashboard

## 🔧 **Implementation Details**

### Backend Changes

#### **Database Migration**
```sql
-- Migration: 006_add_password_changed_field.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed BOOLEAN DEFAULT FALSE;
UPDATE users SET password_changed = TRUE WHERE password_changed IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_password_changed ON users(password_changed);
```

#### **User Struct Update**
```go
type User struct {
    // ... existing fields ...
    PasswordChanged bool // New field
}
```

#### **New Database Functions**
```go
// Update password and mark as changed
func (db *DB) UpdateUserPasswordWithChange(userID int, newPasswordHash string) error

// Regular password update (doesn't change password_changed status)
func (db *DB) UpdateUserPassword(userID int, newPasswordHash string) error
```

#### **Admin Creation Scripts**
- **PostgreSQL**: `cmd/create-admin-postgres/main.go`
- **SQLite**: `cmd/create-admin/main.go`
- **Test Users**: `cmd/create-test-users/main.go`

All scripts now set `password_changed = false` for new users.

### Frontend Changes

#### **New Route**
- **Path**: `/change-password`
- **File**: `src/routes/change-password/+page.svelte`
- **Purpose**: Forced password change form

#### **Route Guard**
- **File**: `src/routes/+layout.ts`
- **Function**: Automatically redirects users to password change if needed

#### **Auth Store Updates**
- **User Interface**: Added `password_changed: boolean`
- **Login Response**: Includes password change status
- **Token Storage**: Preserves password change requirement

## 🎯 **Usage Examples**

### **Creating a Super Admin User**
```bash
cd backend
go run cmd/create-admin-postgres/main.go
```

**Output**:
```
✅ Admin user created successfully with ID: 1
📧 Email: admin@bome.test
🔑 Password: Admin123!
👤 Role: super_admin
✅ Email verified: true
🔒 Password change required on first login
🌐 You can now log in to the admin dashboard
⚠️  IMPORTANT: Change these credentials in production!
🔐 You will be forced to change this password on first login!
```

### **First Login Flow**
1. **Login**: `admin@bome.test` / `Admin123!`
2. **Redirect**: Automatically sent to `/change-password`
3. **Change Password**: Must provide new secure password
4. **Success**: Redirected to admin dashboard
5. **Future Logins**: Normal flow (no password change required)

## 🔒 **Security Features**

### **Password Requirements**
- Minimum 8 characters
- Current password verification
- Password confirmation matching

### **Access Control**
- **Blocked Routes**: All routes except `/change-password`
- **Error Code**: `PASSWORD_CHANGE_REQUIRED`
- **User Experience**: Clear messaging about requirement

### **Audit Logging**
- Password change events logged
- Failed attempts tracked
- Security audit trail maintained

## 🚀 **Deployment Instructions**

### **1. Database Migration**
```bash
# Run the new migration
psql -d your_database -f backend/migrations/006_add_password_changed_field.sql
```

### **2. Create Admin User**
```bash
cd backend
go run cmd/create-admin-postgres/main.go
```

### **3. Deploy Frontend**
- New route `/change-password` automatically included
- Route guards handle redirects
- No additional configuration needed

### **4. Test the Feature**
1. Login with admin credentials
2. Verify redirect to password change
3. Change password successfully
4. Confirm normal access restored

## 🔍 **Troubleshooting**

### **Common Issues**

#### **User Stuck on Password Change Page**
- Check database: `SELECT password_changed FROM users WHERE email = 'user@example.com';`
- Verify password change was successful
- Check audit logs for errors

#### **Password Change Not Working**
- Verify API endpoint `/api/v1/auth/change-password`
- Check authentication middleware
- Validate password requirements

#### **Route Guard Not Working**
- Check browser console for errors
- Verify `+layout.ts` file exists
- Confirm SvelteKit routing configuration

### **Debug Commands**
```bash
# Check user password status
psql -d your_database -c "SELECT email, password_changed FROM users WHERE email = 'admin@bome.test';"

# Force password change (emergency only)
psql -d your_database -c "UPDATE users SET password_changed = TRUE WHERE email = 'admin@bome.test';"
```

## 📚 **API Endpoints**

### **Change Password**
```
POST /api/v1/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "current_password": "Admin123!",
  "new_password": "NewSecurePassword123!"
}
```

### **Response**
```json
{
  "message": "Password changed successfully"
}
```

## 🔮 **Future Enhancements**

### **Planned Features**
- **Password History**: Prevent reuse of recent passwords
- **Password Expiry**: Force change after time period
- **Strength Indicators**: Visual password strength feedback
- **Multi-Factor**: Additional security layers

### **Configuration Options**
- **Password Requirements**: Configurable complexity rules
- **Change Frequency**: Time-based password expiration
- **Admin Override**: Emergency password reset capabilities

## 📞 **Support**

For issues or questions about this feature:
1. Check the troubleshooting section above
2. Review database logs and audit trails
3. Verify configuration and environment variables
4. Contact the development team

---

**Last Updated**: August 2025  
**Version**: 1.0.0  
**Status**: Production Ready ✅
