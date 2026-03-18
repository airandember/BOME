# 🧬 Authentication Braid - Frontend
**Svelte5 Presentation Layer for Authentication System**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **frontend portion** of the Authentication Braid.  
> **Backend portion**: See `_backend/authentication/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Frontend Overview**

**Purpose**: User interface and state management for authentication  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/login`, `/register`, `/auth/*`  
**State Management**: Svelte stores with reactive state

---

## 🌐 **Network Layer: Presentation (Layer 1)**

```
📁 _frontend/authentication/
├── 🧬 BRAID.md                      # This file (frontend overview)
├── layers/
│   └── presentation/                 # SVELTE5 PRESENTATION LAYER
│       ├── 🔗 ELASTIC-BAND-DOWN.md  # ↓ Connection to Backend API
│       ├── pages/
│       │   ├── login-page.md        # → frontend/src/routes/login/+page.svelte
│       │   ├── register-page.md     # → frontend/src/routes/register/+page.svelte
│       │   ├── verify-email.md      # → frontend/src/routes/auth/verify-email/+page.svelte
│       │   └── setup-password.md    # → frontend/src/routes/auth/setup-password/+page.svelte
│       ├── components/
│       │   ├── navigation.md        # → frontend/src/lib/components/Navigation.svelte
│       │   ├── auth-forms.md        # Login/Register form components
│       │   └── oauth2-buttons.md    # OAuth2 login buttons
│       └── stores/
│           ├── auth-store.md        # → frontend/src/lib/auth.ts
│           └── user-store.md        # User state management
└── strands/
    ├── user-registration/           # Frontend portion of registration
    ├── user-login/                  # Frontend portion of login
    ├── email-verification/          # Frontend portion of verification
    ├── session-management/          # Frontend portion of sessions
    └── oauth2-integration/          # Frontend portion of OAuth2
```

---

## 🔗 **Connection to Backend**

### **Elastic Band: Frontend ↔ Backend**
```
🎨 Frontend Presentation (Svelte5)
    ↕️ ELASTIC-BAND-DOWN.md
🔗 Backend Application (REST API)
    ↕️ ELASTIC-BAND-DOWN.md
⚙️  Backend Business Logic (Go)
    ↕️ ELASTIC-BAND-DOWN.md
🗄️  Backend Data Access (Go)
    ↕️ ELASTIC-BAND-DOWN.md
📊 Backend Persistence (PostgreSQL)
```

**Contract File**: `layers/presentation/ELASTIC-BAND-DOWN.md`  
**Connects To**: `_backend/authentication/layers/application/ELASTIC-BAND-UP.md`

---

## 📄 **Frontend Pages**

### **1. Login Page** (`/login`)
**File**: `frontend/src/routes/login/+page.svelte`  
**Purpose**: Email/password login form  
**Features**:
- Email/password input
- Show/hide password toggle
- Error message display
- Loading states
- Forgot password link
- OAuth2 login buttons
- Link to registration

**API Calls**:
- `POST /api/v1/auth/login`

**Redirects**:
- Success → `/dashboard`
- Email not verified → `/auth/verify-email`

---

### **2. Registration Page** (`/register`)
**File**: `frontend/src/routes/register/+page.svelte`  
**Purpose**: New user registration  
**Features**:
- Email, first name, last name inputs
- Real-time email validation
- Error message display
- Loading states
- Link to login
- Privacy policy acceptance

**API Calls**:
- `POST /api/v1/auth/register`

**Redirects**:
- Success → `/auth/verify-email?email={email}&user_id={user_id}`

---

### **3. Email Verification Page** (`/auth/verify-email`)
**File**: `frontend/src/routes/auth/verify-email/+page.svelte`  
**Purpose**: Prompt user to check email  
**Features**:
- Instructions to check email
- Resend verification button
- Email address display
- Countdown timer for resend
- Help section

**API Calls**:
- `POST /api/v1/auth/resend-verification`

**Backend Handler**:
- `GET /api/v1/auth/verify-email-link?token={token}&user_id={user_id}` (from email)

---

### **4. Password Setup Page** (`/auth/setup-password`)
**File**: `frontend/src/routes/auth/setup-password/+page.svelte`  
**Purpose**: Set password after email verification  
**Features**:
- Password input with strength indicator
- Confirm password input
- Password requirements checklist
- Show/hide password toggle
- Auto-login after setup
- Error message display

**API Calls**:
- `POST /api/v1/auth/setup-password`

**Redirects**:
- Success → `/dashboard` (with auto-login)

---

### **5. Forgot Password Page** (`/auth/forgot-password`)
**File**: `frontend/src/routes/auth/forgot-password/+page.svelte`  
**Purpose**: Request password reset  
**Features**:
- Email input
- Success message (always shown for security)
- Link back to login

**API Calls**:
- `POST /api/v1/auth/forgot-password`

---

### **6. Reset Password Page** (`/auth/reset-password`)
**File**: `frontend/src/routes/auth/reset-password/+page.svelte`  
**Purpose**: Reset password with token  
**Features**:
- Password input with strength indicator
- Confirm password input
- Token validation
- Success message

**API Calls**:
- `POST /api/v1/auth/reset-password`

**Redirects**:
- Success → `/login`

---

## 🧩 **Frontend Components**

### **Navigation Component**
**File**: `frontend/src/lib/components/Navigation.svelte`  
**Purpose**: Auth-aware navigation  
**Features**:
- Shows different links based on auth state
- Login/Logout buttons
- User profile dropdown
- Role-based menu items

**Uses**: `$auth.isAuthenticated`, `$auth.user`

---

### **Auth Form Components**
**Files**: Various form components  
**Purpose**: Reusable form elements  
**Features**:
- Email input with validation
- Password input with show/hide
- Form error display
- Loading button states
- Form field validation

---

### **OAuth2 Buttons**
**Purpose**: Social login buttons  
**Features**:
- Google login button
- GitHub login button (if enabled)
- Styled with brand colors
- Loading states

**API Calls**:
- `GET /api/v1/auth/oauth2/{provider}/login`

---

## 🗃️ **Frontend Stores**

### **Auth Store** (`$lib/auth.ts`)
**Purpose**: Central authentication state management  
**State**:
```typescript
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
}
```

**Methods**:
- `login(email, password)` → Login user
- `register(data)` → Register new user
- `logout()` → Logout user
- `refreshToken()` → Refresh JWT
- `getCurrentUser()` → Get current user
- `forgotPassword(email)` → Request reset
- `resetPassword(token, password)` → Reset password

**Storage**:
- JWT tokens in `SecureTokenStorage`
- User data in Svelte store
- Reactive across all components

---

## 🔄 **Frontend Data Flow**

### **Login Flow**:
```
1. User fills form → /login
2. Component validates input
3. Calls auth.login(email, password)
4. Store calls API: POST /api/v1/auth/login
5. Backend validates & returns JWT
6. Store saves tokens to SecureTokenStorage
7. Store updates state: isAuthenticated = true
8. Component redirects to /dashboard
```

### **Registration Flow**:
```
1. User fills form → /register
2. Component validates input
3. Calls auth.register(data)
4. Store calls API: POST /api/v1/auth/register
5. Backend creates user & sends email
6. Component redirects to /auth/verify-email
7. User clicks link in email
8. Redirects to /auth/setup-password
9. User sets password
10. Auto-login & redirect to /dashboard
```

---

## 🔒 **Frontend Security**

### **Token Management**:
- Tokens stored in localStorage (future: httpOnly cookies)
- Automatic token refresh before expiration
- Tokens cleared on logout
- Secure token transmission (HTTPS only)

### **XSS Prevention**:
- Svelte auto-escapes all output
- No use of `{@html}` with user input
- Content Security Policy headers

### **CSRF Protection**:
- JWT tokens provide CSRF protection
- No additional CSRF token needed

### **Input Validation**:
- Client-side validation before API calls
- Server-side validation as source of truth
- Sanitize all user input

---

## ⚡ **Performance Optimization**

### **Code Splitting**:
- Auth routes lazy-loaded
- Components loaded on-demand
- Minimal initial bundle

### **State Management**:
- Reactive Svelte stores
- Minimal re-renders
- Efficient reactivity

### **API Calls**:
- Automatic token refresh
- Request deduplication
- Loading state management

---

## 🎨 **UI/UX Patterns**

### **Loading States**:
```svelte
{#if loading}
  <button disabled>
    <Spinner /> Logging in...
  </button>
{:else}
  <button>Log In</button>
{/if}
```

### **Error Display**:
```svelte
{#if error}
  <div class="error-message">
    <AlertIcon />
    <p>{error}</p>
  </div>
{/if}
```

### **Success Messages**:
```svelte
{#if success}
  <div class="success-message">
    <CheckIcon />
    <p>{successMessage}</p>
  </div>
{/if}
```

### **Form Validation**:
```svelte
<input
  bind:value={email}
  class:invalid={email && !isValidEmail(email)}
/>
{#if email && !isValidEmail(email)}
  <p class="validation-error">Please enter a valid email</p>
{/if}
```

---

## 📊 **Frontend Analytics**

**Events Tracked**:
- `login_attempt` - User clicked login
- `login_success` - Login succeeded
- `login_failed` - Login failed
- `register_attempt` - User clicked register
- `register_success` - Registration succeeded
- `password_reset_requested` - Reset requested
- `oauth2_login_started` - OAuth2 initiated

**Implementation**:
```typescript
import { analytics } from '$lib/integrations/analytics';

auth.login(email, password).then(() => {
  analytics.track('login_success', {
    method: 'email',
    user_id: user.id,
  });
});
```

---

## 🔗 **Related Documentation**

### **Backend Portion of This Braid**:
- [`_braids/authentication/backend/BRAID.md`](../../_braids/authentication/backend/BRAID.md) - Backend overview
- [`_braids/authentication/backend/layers/application/ELASTIC-BAND-UP.md`](../../_braids/authentication/backend/layers/application/ELASTIC-BAND-UP.md) - API contracts

### **Frontend Elastic Band**:
- [`layers/presentation/ELASTIC-BAND-DOWN.md`](layers/presentation/ELASTIC-BAND-DOWN.md) - Frontend → Backend contract

### **Strands** (Both Frontend & Backend):
- [`strands/user-login/`](strands/user-login/) - Complete login flow
- [`strands/user-registration/`](strands/user-registration/) - Complete registration flow
- [`strands/email-verification/`](strands/email-verification/) - Email verification flow
- [`strands/session-management/`](strands/session-management/) - Session handling
- [`strands/oauth2-integration/`](strands/oauth2-integration/) - OAuth2 flow

---

## 🎯 **Quick Links**

**Actual Files**:
- Auth Store: `frontend/src/lib/auth.ts`
- Login Page: `frontend/src/routes/login/+page.svelte`
- Register Page: `frontend/src/routes/register/+page.svelte`
- Verify Email: `frontend/src/routes/auth/verify-email/+page.svelte`
- Setup Password: `frontend/src/routes/auth/setup-password/+page.svelte`
- Navigation: `frontend/src/lib/components/Navigation.svelte`

**API Documentation**:
- See `_braids/authentication/backend/layers/business-logic/ELASTIC-BAND-UP.md`

---

## 🐛 **Common Frontend Issues**

### **Issue: "Network Error"**
**Cause**: Backend not running or CORS issue  
**Solution**: Check backend is running on correct port, verify CORS settings

### **Issue: Tokens not persisting**
**Cause**: localStorage not available or being cleared  
**Solution**: Check browser settings, verify SecureTokenStorage

### **Issue: Infinite redirect loop**
**Cause**: Auth check on protected routes  
**Solution**: Verify auth middleware logic, check token validation

### **Issue: Form not submitting**
**Cause**: Validation errors or loading state  
**Solution**: Check browser console, verify form validation

---

## 📝 **Development Guidelines**

### **Adding New Auth Pages**:
1. Create page in `frontend/src/routes/auth/`
2. Document in `_frontend/braids/authentication/layers/presentation/pages/`
3. Add to navigation if needed
4. Update this BRAID.md

### **Modifying Auth Store**:
1. Update store in `frontend/src/lib/auth.ts`
2. Update documentation in `layers/presentation/stores/auth-store.md`
3. Update elastic band contract
4. Test all consuming components

### **Adding API Calls**:
1. Define API contract in backend elastic band
2. Implement in auth store
3. Update frontend elastic band
4. Add error handling

---

**Last Updated**: October 14, 2025  
**Status**: Complete frontend documentation  
**Technology**: Svelte 5 + TypeScript  
**Backend Counterpart**: `_braids/authentication/backend/`

---

**Navigate**:  
[🏠 Master Index](../../../BRAIDS_INDEX.md) | [⬅️ Backend Braid](../../_braids/authentication/backend/BRAID.md) | [📚 Getting Started](../../_braids/authentication/backend/GETTING_STARTED.md)

