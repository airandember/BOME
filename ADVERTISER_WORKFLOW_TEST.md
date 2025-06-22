# Advertiser Workflow Testing Guide

This guide explains how to test the complete advertiser registration and approval workflow in BOME.

## Overview

The workflow involves three main components:
1. **User Registration** (`user@bome.test`) - Submit advertiser application
2. **Admin Approval** (`admin@bome.test`) - Review and approve/reject applications  
3. **User Dashboard** (`user@bome.test`) - View application status and updates

## Testing Steps

### Step 1: User Registration
1. Open the application at `http://localhost:5174`
2. Login as `user@bome.test` / `password123`
3. Navigate to `/advertise` or click "Advertise with BOME" 
4. Fill out the business registration form:
   - Company Name: "Test Business Co"
   - Business Email: "business@testcompany.com"
   - Contact Name: "John Doe"
   - Contact Phone: "+1 (555) 123-4567"
   - Industry: Select any option
   - (Optional fields can be left blank)
5. Click "Continue to Package Selection"
6. Choose any package (e.g., "Professional")
7. The application will be submitted and you'll see the success page
8. Click "Return to Dashboard" - you'll be redirected to the dashboard's Advertiser tab
9. **Expected Result**: Dashboard shows "Application Under Review" with pending status

### Step 2: Admin Review
1. Logout and login as `admin@bome.test` / `admin123`
2. Navigate to `/admin/advertisements`
3. You should see the new advertiser application in the "Pending" section
4. **To Approve**:
   - Click the "Approve" button next to the application
   - The application will move to the "Approved" section
5. **To Reject**:
   - Click the "Reject" button next to the application
   - Enter a rejection reason when prompted (e.g., "Incomplete documentation")
   - The application will move to the "Rejected" section

### Step 3: User Dashboard Updates
1. Logout and login back as `user@bome.test` / `password123`
2. Navigate to `/dashboard` and click the "Advertiser Information" tab
3. **If Approved**: 
   - Dashboard shows "Welcome, Approved Advertiser!" 
   - Account overview with campaign metrics
   - Quick actions to create campaigns
4. **If Rejected**:
   - Dashboard shows "Application Not Approved"
   - Displays the rejection reason provided by admin
   - Shows steps to reapply

## Technical Implementation

### Data Flow
- **Advertiser Store** (`/src/lib/stores/advertiser.ts`) manages all advertiser account state
- **Real-time Updates** - Changes made by admin are immediately reflected in user dashboard
- **Persistent State** - Application status persists across browser sessions

### Key Components
- **Registration Form** (`/src/routes/advertise/+page.svelte`)
- **Admin Dashboard** (`/src/routes/admin/advertisements/+page.svelte`) 
- **User Dashboard** (`/src/routes/dashboard/+page.svelte`)
- **Advertiser Store** (`/src/lib/stores/advertiser.ts`)

### Mock Data
- Initial mock account exists for user_id: 2
- Admin actions update the shared store
- All changes are reflected across components immediately

## Testing Different Scenarios

### Scenario 1: Complete Approval Flow
1. Register as new advertiser → Submit application → Admin approves → User sees approved status

### Scenario 2: Rejection and Reapplication  
1. Register as new advertiser → Submit application → Admin rejects → User sees rejection → User can reapply

### Scenario 3: Existing Advertiser
1. If user already has an advertiser account, they're redirected to dashboard instead of registration

## Notes
- The system uses mock data and simulated API delays for realistic testing
- In production, this would connect to real backend APIs
- All state changes are immediate and persistent within the session
- The workflow supports the complete advertiser lifecycle from registration to approval/rejection 