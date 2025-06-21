# BOME Test Accounts

This document provides information about test accounts available for the BOME (Book of Mormon Evidence) platform.

## Available Test Accounts

### 1. Admin Account (Super Administrator)
- **Email**: `admin@bome.com` or `admin@bome.test`
- **Password**: `admin123`
- **Role**: Super Administrator
- **Access**: Full admin dashboard access with all permissions
- **Features**: 
  - User management
  - Content management
  - Advertisement management
  - Financial management
  - System administration
  - Role management
  - Analytics and reporting

### 2. Regular User Account
- **Email**: `user@bome.com` or `user@bome.test`
- **Password**: `user123`
- **Role**: Regular User
- **Access**: User-level dashboard and features
- **Features**:
  - Video streaming access
  - User dashboard
  - Account settings
  - Subscription management
  - Favorites and playlists
  - Comment system

### 3. Advertiser Account
- **Email**: `advertiser@bome.com` or `advertiser@bome.test`
- **Password**: `advertiser123`
- **Role**: Advertiser
- **Access**: Advertiser dashboard and campaign management
- **Features**:
  - Advertiser dashboard
  - Campaign creation and management
  - Advertisement creation and management
  - Analytics and reporting
  - Budget and billing management
  - Asset management
  - Performance tracking

### 4. Business Advertiser Account
- **Email**: `business@bome.test`
- **Password**: `business123`, `advertiser123`, or `password123`
- **Role**: Advertiser
- **Access**: Advertiser dashboard and campaign management
- **Features**:
  - Advertiser dashboard
  - Campaign creation and management
  - Advertisement creation and management
  - Analytics and reporting
  - Budget and billing management
  - Asset management
  - Performance tracking

## How to Use Test Accounts

1. **Start the Development Server**:
   ```bash
   cd frontend && npm run dev
   ```

2. **Navigate to Login Page**:
   - Go to `http://localhost:5176/login` (or whatever port Vite assigns)

3. **Login with Test Credentials**:
   - Use any of the email/password combinations above
   - The system will automatically authenticate and redirect appropriately

4. **Access Different Dashboards**:
   - **Admin**: Will be redirected to `/admin` with full admin interface
   - **Advertiser**: Will be redirected to `/advertiser` with advertiser dashboard
   - **Regular User**: Will be redirected to `/dashboard` with user interface

## Testing Different User Experiences

### Admin Testing
- Navigate to `/admin` to access the admin dashboard
- Test advertisement management at `/admin/advertisements`
- Test user management at `/admin/users`
- Test role management at `/admin/roles`
- Test financial management at `/admin/financial`

### Advertiser Testing
- Navigate to `/advertiser` for advertiser dashboard
- Test campaign creation at `/advertiser/campaigns/new`
- Test campaign management at `/advertiser/campaigns`
- Test advertisement creation at `/advertiser/campaigns/[id]/ads/new`
- Test analytics dashboard at `/advertiser/analytics`
- Test account management at `/advertiser/account`
- Test account setup at `/advertiser/setup` (if not already set up)

### Regular User Testing
- Navigate to `/dashboard` for user dashboard
- Test video streaming at `/videos`
- Test subscription management at `/subscription`
- Test account settings at `/account`
- Test billing at `/account/billing`

## Mock Data Available

The test accounts come with pre-populated mock data including:
- **Videos**: 25 sample videos with Bunny.net streaming URLs
- **Articles**: 18 comprehensive articles with categories and tags
- **Events**: 6 sample events with different types and statuses
- **Advertisements**: Sample advertiser accounts and campaigns
- **Analytics**: Mock analytics data for dashboards
- **Roles**: 32 predefined roles with different permission levels

## Notes

- These are mock accounts for development and testing only
- No real authentication is performed - credentials are hardcoded
- Data is stored in localStorage and will persist across sessions
- To clear session data, use the logout function or clear browser localStorage
- The backend API endpoints are mocked and return sample data

## Switching Between Accounts

To test different user experiences:
1. Logout from current account
2. Login with different test credentials
3. Navigate to appropriate dashboard sections

## Asset Testing

The system includes test assets in `/frontend/src/lib/HOMEPAGE_TEST_ASSETS/` including:
- Placeholder images (16x10 format)
- Historical maps (1830s New York, United States)
- World maps and globes
- Book of Mormon imagery
- Various research documents

These assets can be used for testing image uploads, content management, and advertisement asset functionality. 