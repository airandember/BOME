# 🚀 Quick Test Reference Card

## 🌐 Access URLs
- **Frontend**: http://localhost:5173
- **Backend**: http://localhost:8080

## 🔑 Essential Test Users

### 👑 **Super Administrator** (Full Access)
```
Email: super.admin@bome.test
Password: SuperAdmin123!
```
**Test**: Access all dashboards and functions

### 📝 **Content Manager** (Content Management)
```
Email: content.manager@bome.test
Password: ContentManager123!
```
**Test**: Manage articles, videos, and content

### 📊 **Analytics Manager** (Analytics Access)
```
Email: analytics.manager@bome.test
Password: AnalyticsManager123!
```
**Test**: View all analytics and reports

### 💰 **Financial Administrator** (Financial Access)
```
Email: financial.admin@bome.test
Password: FinancialAdmin123!
```
**Test**: Access billing and financial data

### 👥 **User Manager** (User Management)
```
Email: user.manager@bome.test
Password: UserManager123!
```
**Test**: Manage user accounts and support

### 🎯 **Advertiser** (Advertiser Dashboard)
```
Email: advertiser@bome.test
Password: Advertiser123!
```
**Test**: Create and manage ad campaigns

### 👤 **Basic User** (Limited Access)
```
Email: user@bome.test
Password: User123!
```
**Test**: Basic platform access only

## 🧪 Quick Test Scenarios

### 1. **Role Hierarchy Test**
1. Login as `super.admin@bome.test` → Should see all options
2. Login as `user@bome.test` → Should see limited options
3. Compare dashboard access differences

### 2. **Subsystem Access Test**
1. Login as `articles.manager@bome.test` → Articles access only
2. Login as `youtube.manager@bome.test` → YouTube access only
3. Verify cross-subsystem restrictions

### 3. **Permission Test**
1. Login as `content.creator@bome.test` → Can create content
2. Login as `content.manager@bome.test` → Can approve content
3. Test approval workflow

## ⚡ 5-Minute Test Plan

1. **Login as Super Admin** → Verify full access
2. **Login as Content Manager** → Test content functions
3. **Login as Advertiser** → Test advertiser dashboard
4. **Login as Basic User** → Verify limited access
5. **Test role switching** → Verify proper access control

## 🚨 Common Issues

- **Can't login?** → Check if servers are running
- **Wrong access?** → Check user role in database
- **Missing features?** → Check role permissions

## 📞 Need Help?

- Check `TEST_USERS_GUIDE.md` for detailed scenarios
- Verify database connection and migrations
- Check server logs for errors 