# 🧪 BOME Test Users Guide

## 🎯 Overview
This guide provides test credentials and scenarios for testing the standardized role system across different dashboards and subsystems.

## 🚀 Quick Start
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **Database**: PostgreSQL (bome_db)

---

## 👥 Test Users by Role Level

### 🔴 **Level 10: Super Administrator**
**Full system access and role management capabilities**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `super.admin@bome.test` | `SuperAdmin123!` | Super Administrator | Super Administrator | **Full system access** |

**Test Scenarios:**
- ✅ Access all admin dashboards
- ✅ Manage all users and roles
- ✅ View all analytics and reports
- ✅ Access system settings and security
- ✅ Manage all content and subsystems

---

### 🟠 **Level 9: System Administrator**
**Technical system management without role changes**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `system.admin@bome.test` | `SystemAdmin123!` | System Administrator | System Administrator | **Technical management** |

**Test Scenarios:**
- ✅ Access system health and monitoring
- ✅ Manage security settings
- ✅ View technical logs and backups
- ✅ Configure integrations
- ❌ Cannot manage user roles

---

### 🟡 **Level 8: Content Manager**
**Overall content strategy and oversight**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `content.manager@bome.test` | `ContentManager123!` | Sarah Johnson | Content Manager | **Content oversight** |

**Test Scenarios:**
- ✅ Manage all content (articles, videos, streaming)
- ✅ Approve and publish content
- ✅ View content analytics
- ✅ Moderate user-generated content
- ❌ Cannot access system administration

---

### 🟢 **Level 7: Subsystem & Department Managers**
**Specialized management roles**

| Email | Password | Name | Role | Subsystem |
|-------|----------|------|------|-----------|
| `articles.manager@bome.test` | `ArticlesManager123!` | Emily Rodriguez | Articles Manager | Articles |
| `youtube.manager@bome.test` | `YouTubeManager123!` | Alex Kim | YouTube Manager | YouTube |
| `streaming.manager@bome.test` | `StreamingManager123!` | Jessica Wang | Video Streaming Manager | Streaming |
| `events.manager@bome.test` | `EventsManager123!` | Robert Davis | Events Manager | Events |
| `advertisement.manager@bome.test` | `AdManager123!` | Lisa Brown | Advertisement Manager | Marketing |
| `user.manager@bome.test` | `UserManager123!` | Rachel Green | User Account Manager | User Management |
| `analytics.manager@bome.test` | `AnalyticsManager123!` | Amanda Taylor | Analytics Manager | Analytics |
| `financial.admin@bome.test` | `FinancialAdmin123!` | Daniel Anderson | Financial Administrator | Financial |

**Test Scenarios:**
- ✅ Manage their specific subsystem
- ✅ View subsystem analytics
- ✅ Approve content/campaigns in their area
- ❌ Cannot access other subsystems
- ❌ Cannot manage system settings

---

### 🔵 **Level 6: Content Creator & Academic Reviewer**
**Content creation and academic review**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `content.creator@bome.test` | `ContentCreator123!` | David Thompson | Content Creator | Content creation |
| `academic.reviewer@bome.test` | `AcademicReviewer123!` | Dr. Rebecca Williams | Academic Reviewer | Academic review |

**Test Scenarios:**
- ✅ Create and edit content
- ✅ Review academic content for accuracy
- ✅ Submit content for approval
- ❌ Cannot publish without approval
- ❌ Cannot access management functions

---

### 🟣 **Level 5: Support & Technical Specialists**
**Specialized support roles**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `support.specialist@bome.test` | `SupportSpecialist123!` | Kevin Martinez | Support Specialist | User support |
| `technical.specialist@bome.test` | `TechnicalSpecialist123!` | Chris Lee | Technical Specialist | Technical support |
| `research.coordinator@bome.test` | `ResearchCoordinator123!` | Dr. James Miller | Research Coordinator | Research coordination |

**Test Scenarios:**
- ✅ Provide user support
- ✅ Access user accounts for assistance
- ✅ Coordinate research activities
- ❌ Cannot manage system settings
- ❌ Cannot approve content

---

### 🟤 **Level 4: Marketing Specialist**
**Marketing and campaign management**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `marketing.specialist@bome.test` | `MarketingSpecialist123!` | Tom Wilson | Marketing Specialist | Marketing campaigns |

**Test Scenarios:**
- ✅ Create marketing campaigns
- ✅ Manage advertiser relationships
- ✅ View campaign analytics
- ❌ Cannot approve campaigns
- ❌ Cannot access financial data

---

### 🟡 **Level 3: Advertiser**
**Advertiser account access**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `advertiser@bome.test` | `Advertiser123!` | Maria Garcia | Advertiser | Advertiser dashboard |

**Test Scenarios:**
- ✅ Create advertising campaigns
- ✅ View campaign performance
- ✅ Manage billing and payments
- ❌ Cannot access admin functions
- ❌ Cannot view other advertisers' data

---

### ⚪ **Level 1: Basic User**
**Standard platform access**

| Email | Password | Name | Role | Access |
|-------|----------|------|------|--------|
| `user@bome.test` | `User123!` | John Doe | User | Basic platform access |

**Test Scenarios:**
- ✅ View public content
- ✅ Create user account
- ✅ Access basic features
- ❌ Cannot access admin functions
- ❌ Cannot create campaigns

---

## 🧪 Testing Scenarios

### **1. Role Hierarchy Testing**
Test that higher-level roles can access lower-level functions:

1. Login as `super.admin@bome.test`
2. Verify access to all dashboards
3. Login as `content.manager@bome.test`
4. Verify access to content management only
5. Login as `user@bome.test`
6. Verify limited access

### **2. Subsystem Access Testing**
Test subsystem-specific access:

1. Login as `articles.manager@bome.test`
2. Verify access to articles dashboard
3. Verify NO access to YouTube or streaming dashboards
4. Login as `youtube.manager@bome.test`
5. Verify access to YouTube dashboard only

### **3. Permission Inheritance Testing**
Test that permissions are properly inherited:

1. Login as `super.admin@bome.test`
2. Verify all permissions are available
3. Login as `content.creator@bome.test`
4. Verify only content creation permissions

### **4. Cross-Subsystem Access Testing**
Test that users cannot access unauthorized subsystems:

1. Login as `advertisement.manager@bome.test`
2. Try to access articles management
3. Verify access is denied
4. Try to access user management
5. Verify access is denied

### **5. Content Approval Workflow Testing**
Test content approval process:

1. Login as `content.creator@bome.test`
2. Create new content
3. Submit for approval
4. Login as `content.manager@bome.test`
5. Approve the content
6. Verify content is published

---

## 🔍 Dashboard Testing Checklist

### **Admin Dashboard** (`/admin`)
- [ ] Super Admin: Full access to all sections
- [ ] System Admin: Access to system, security, technical sections
- [ ] Content Manager: Access to content, analytics sections
- [ ] Other roles: Limited or no access

### **Content Management** (`/admin/videos`, `/admin/articles`)
- [ ] Content Manager: Full access
- [ ] Content Editor: Edit and publish access
- [ ] Content Creator: Create and edit access
- [ ] Other roles: No access

### **User Management** (`/admin/users`)
- [ ] Super Admin: Full access
- [ ] User Manager: Full access
- [ ] Support Specialist: Limited access
- [ ] Other roles: No access

### **Analytics Dashboard** (`/admin/analytics`)
- [ ] Super Admin: Full access
- [ ] Analytics Manager: Full access
- [ ] Content Manager: Content analytics only
- [ ] Other roles: Limited or no access

### **Financial Dashboard** (`/admin/financial`)
- [ ] Super Admin: Full access
- [ ] Financial Admin: Full access
- [ ] Advertisement Manager: Ad revenue only
- [ ] Other roles: No access

### **Advertiser Dashboard** (`/advertiser`)
- [ ] Advertiser: Full access to own campaigns
- [ ] Advertisement Manager: Access to all campaigns
- [ ] Marketing Specialist: Limited access
- [ ] Other roles: No access

---

## 🚨 Security Testing

### **1. Unauthorized Access Prevention**
- [ ] Users cannot access dashboards above their role level
- [ ] Users cannot access subsystems they're not authorized for
- [ ] API endpoints properly validate permissions

### **2. Data Isolation**
- [ ] Users can only see their own data
- [ ] Advertisers cannot see other advertisers' campaigns
- [ ] Content creators cannot see unpublished content from others

### **3. Role Escalation Prevention**
- [ ] Users cannot modify their own roles
- [ ] Users cannot grant themselves higher permissions
- [ ] Role changes require appropriate authorization

---

## 📊 Expected Results

### **Access Matrix**

| Role Level | Admin | Content | Analytics | Financial | User Mgmt | Security |
|------------|-------|---------|-----------|-----------|-----------|----------|
| Super Admin (10) | ✅ Full | ✅ Full | ✅ Full | ✅ Full | ✅ Full | ✅ Full |
| System Admin (9) | ❌ | ❌ | ✅ Read | ❌ | ❌ | ✅ Full |
| Content Manager (8) | ❌ | ✅ Full | ✅ Content | ❌ | ❌ | ❌ |
| Subsystem Managers (7) | ❌ | ✅ Subsystem | ✅ Subsystem | ❌ | ❌ | ❌ |
| Content Creator (6) | ❌ | ✅ Create | ❌ | ❌ | ❌ | ❌ |
| Support (5) | ❌ | ❌ | ❌ | ❌ | ✅ Limited | ❌ |
| Marketing (4) | ❌ | ❌ | ✅ Marketing | ❌ | ❌ | ❌ |
| Advertiser (3) | ❌ | ❌ | ✅ Own | ✅ Own | ❌ | ❌ |
| User (1) | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## 🎯 Next Steps

1. **Test each user login** and verify dashboard access
2. **Test role-based functionality** within each dashboard
3. **Test permission inheritance** across role levels
4. **Test cross-subsystem access** restrictions
5. **Test content approval workflows**
6. **Test security boundaries** and data isolation

---

## ⚠️ Important Notes

- **These are test credentials only** - do not use in production
- **All passwords follow the pattern**: `RoleName123!`
- **All emails use the domain**: `@bome.test`
- **Database is PostgreSQL** with standardized roles
- **Frontend and backend must be running** for full testing

---

## 🆘 Troubleshooting

### **Login Issues**
- Verify both frontend (5173) and backend (8080) are running
- Check database connection
- Verify user exists in database

### **Access Issues**
- Check user role assignments in `user_roles` table
- Verify role permissions in `role_permissions` table
- Check frontend role validation logic

### **Database Issues**
- Verify PostgreSQL is running
- Check database connection string
- Verify migration was applied successfully 