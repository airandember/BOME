# 🔗 ELASTIC BAND: Presentation → Application
**Interface Contract Between Svelte5 Frontend and Backend REST API**

---

## 📍 **Connection Points**

**From**: Svelte5 Frontend Components & Stores (Layer 1 - Presentation)  
**To**: Go Backend REST API (Layer 2 - Application)  
**Purpose**: Define how frontend consumes backend authentication API

> **⚠️ CROSS-REPOSITORY CONTRACT**  
> **This file**: `_braids/authentication/frontend/layers/presentation/ELASTIC-BAND-DOWN.md`  
> **Connects to**: `_braids/authentication/backend/layers/application/ELASTIC-BAND-UP.md`

---

## 🌐 **API Base Configuration**

### **Environment Configuration**:
```typescript
// frontend/.env.development
VITE_API_URL=http://localhost:8080

// frontend/.env.production
VITE_API_URL=https://api.yourdomain.com
```

### **API Client Setup**:
```typescript
// frontend/src/lib/api/client.ts
export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
export const API_VERSION = 'v1';
export const API_URL = `${API_BASE_URL}/api/${API_VERSION}`;

// All auth endpoints
export const AUTH_ENDPOINTS = {
  LOGIN: `${API_URL}/auth/login`,
  REGISTER: `${API_URL}/auth/register`,
  LOGOUT: `${API_URL}/auth/logout`,
  REFRESH: `${API_URL}/auth/refresh`,
  ME: `${API_URL}/auth/me`,
  VERIFY_EMAIL: `${API_URL}/auth/verify-email-link`,
  SETUP_PASSWORD: `${API_URL}/auth/setup-password`,
  RESEND_VERIFICATION: `${API_URL}/auth/resend-verification`,
  FORGOT_PASSWORD: `${API_URL}/auth/forgot-password`,
  RESET_PASSWORD: `${API_URL}/auth/reset-password`,
  OAUTH2_LOGIN: (provider: string) => `${API_URL}/auth/oauth2/${provider}/login`,
  OAUTH2_CALLBACK: (provider: string) => `${API_URL}/auth/oauth2/${provider}/callback`,
};
```

---

## 🎯 **Frontend → Backend API Calls**

### **1. User Login**

**Frontend Call** (`frontend/src/lib/auth.ts`):
```typescript
async login(email: string, password: string): Promise<LoginResponse> {
  const response = await fetch(AUTH_ENDPOINTS.LOGIN, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw {
      code: error.code,
      message: error.error,
      status: response.status,
    };
  }

  const result: LoginResponse = await response.json();
  
  // Store tokens
  SecureTokenStorage.setTokens({
    accessToken: result.access_token,
    refreshToken: result.refresh_token,
    expiresAt: Date.now() + (result.expires_in * 1000),
  });
  
  return result;
}
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 14400,
  "user": {
    "id": "123",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": true,
    "role": "user"
  }
}
```

**Error Response** (401 Unauthorized):
```json
{
  "error": "Invalid email or password",
  "code": "INVALID_CREDENTIALS"
}
```

**Backend Handler**: `POST /api/v1/auth/login`  
**Backend File**: `backend/internal/routes/auth.go:LoginHandler()`  
**See**: `_braids/authentication/backend/layers/business-logic/ELASTIC-BAND-UP.md`

---

### **2. User Registration**

**Frontend Call**:
```typescript
async register(data: RegisterData): Promise<RegisterResponse> {
  const response = await fetch(AUTH_ENDPOINTS.REGISTER, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email: data.email,
      first_name: data.first_name,
      last_name: data.last_name,
    }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw {
      code: error.code,
      message: error.error,
      status: response.status,
    };
  }

  return await response.json();
}
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Success Response** (201 Created):
```json
{
  "success": true,
  "message": "Verification email sent. Please check your inbox.",
  "user_id": "123",
  "email": "user@example.com"
}
```

**Error Response** (409 Conflict):
```json
{
  "error": "Email already exists",
  "code": "DUPLICATE_EMAIL"
}
```

**Backend Handler**: `POST /api/v1/auth/register`  
**Backend File**: `backend/internal/routes/auth.go:RegisterHandler()`

---

### **3. Token Refresh**

**Frontend Call**:
```typescript
async refreshToken(): Promise<TokenResponse> {
  const tokens = SecureTokenStorage.getTokens();
  
  const response = await fetch(AUTH_ENDPOINTS.REFRESH, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      refresh_token: tokens.refreshToken,
    }),
  });

  if (!response.ok) {
    // Refresh failed, logout user
    this.logout();
    throw new Error('Session expired. Please log in again.');
  }

  const result: TokenResponse = await response.json();
  
  // Update stored tokens
  SecureTokenStorage.setTokens({
    accessToken: result.access_token,
    refreshToken: result.refresh_token,
    expiresAt: Date.now() + (result.expires_in * 1000),
  });
  
  return result;
}
```

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 14400
}
```

**Backend Handler**: `POST /api/v1/auth/refresh`

---

### **4. Logout**

**Frontend Call**:
```typescript
async logout(): Promise<void> {
  const tokens = SecureTokenStorage.getTokens();
  
  if (tokens?.accessToken) {
    try {
      await fetch(AUTH_ENDPOINTS.LOGOUT, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
      });
    } catch (err) {
      // Logout anyway, even if backend call fails
      console.error('Logout error:', err);
    }
  }
  
  // Clear local state
  SecureTokenStorage.clearTokens();
  authStore.set({
    isAuthenticated: false,
    user: null,
    loading: false,
    error: null,
  });
  
  // Redirect to login
  goto('/login');
}
```

**Request**: No body, just Authorization header

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Backend Handler**: `POST /api/v1/auth/logout`

---

### **5. Get Current User**

**Frontend Call**:
```typescript
async getCurrentUser(): Promise<User> {
  const tokens = SecureTokenStorage.getTokens();
  
  const response = await fetch(AUTH_ENDPOINTS.ME, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${tokens.accessToken}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to get current user');
  }

  const user: User = await response.json();
  
  // Update store
  authStore.update(state => ({
    ...state,
    user,
    isAuthenticated: true,
  }));
  
  return user;
}
```

**Success Response** (200 OK):
```json
{
  "id": "123",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "email_verified": true,
  "role": "user",
  "is_active": true,
  "profile_picture_url": "https://...",
  "created_at": "2025-01-15T10:30:00Z",
  "last_login": "2025-10-14T09:00:00Z"
}
```

**Backend Handler**: `GET /api/v1/auth/me`

---

## 🔒 **Authentication Header**

### **Adding JWT Token to Requests**:
```typescript
// Automatic token injection
async function authenticatedFetch(url: string, options: RequestInit = {}) {
  const tokens = SecureTokenStorage.getTokens();
  
  return fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      ...(tokens?.accessToken && {
        'Authorization': `Bearer ${tokens.accessToken}`,
      }),
    },
  });
}
```

### **Auto-Refresh on 401**:
```typescript
async function authenticatedFetch(url: string, options: RequestInit = {}) {
  const response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${getAccessToken()}`,
    },
  });
  
  // If token expired, try to refresh and retry
  if (response.status === 401) {
    try {
      await auth.refreshToken();
      
      // Retry original request with new token
      return fetch(url, {
        ...options,
        headers: {
          ...options.headers,
          'Authorization': `Bearer ${getAccessToken()}`,
        },
      });
    } catch (refreshError) {
      // Refresh failed, logout
      auth.logout();
      throw refreshError;
    }
  }
  
  return response;
}
```

---

## ⚠️ **Error Handling**

### **Frontend Error Handler**:
```typescript
function handleApiError(error: any): string {
  // Check for specific error codes
  switch (error.code) {
    case 'INVALID_CREDENTIALS':
      return 'Invalid email or password';
    
    case 'EMAIL_NOT_VERIFIED':
      return 'Please verify your email before logging in';
    
    case 'ACCOUNT_SUSPENDED':
      return 'Your account has been suspended';
    
    case 'RATE_LIMIT_EXCEEDED':
      return 'Too many attempts. Please try again later.';
    
    case 'DUPLICATE_EMAIL':
      return 'An account with this email already exists';
    
    case 'WEAK_PASSWORD':
      return 'Password must be at least 8 characters with uppercase, lowercase, and numbers';
    
    case 'PASSWORD_MISMATCH':
      return 'Passwords do not match';
    
    case 'TOKEN_EXPIRED':
      return 'This link has expired. Please request a new one.';
    
    case 'INVALID_TOKEN':
      return 'Invalid verification link';
    
    default:
      // Generic error
      return error.message || 'An error occurred. Please try again.';
  }
}
```

### **Usage in Components**:
```svelte
<script>
  import { auth } from '$lib/auth';
  import { handleApiError } from '$lib/api/errorHandler';
  
  let error = '';
  
  async function handleLogin() {
    try {
      await auth.login(email, password);
    } catch (err) {
      error = handleApiError(err);
    }
  }
</script>

{#if error}
  <div class="error-alert">
    {error}
  </div>
{/if}
```

---

## 🔄 **State Synchronization**

### **Frontend Auth State**:
```typescript
// frontend/src/lib/stores/auth.ts
import { writable, derived } from 'svelte/store';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
}

export const authStore = writable<AuthState>({
  isAuthenticated: false,
  user: null,
  loading: false,
  error: null,
});

// Derived stores
export const isAuthenticated = derived(authStore, $auth => $auth.isAuthenticated);
export const currentUser = derived(authStore, $auth => $auth.user);
export const isAdmin = derived(authStore, $auth => $auth.user?.role === 'admin');
```

### **React to Backend State Changes**:
```typescript
// When user logs in on backend
auth.login(email, password).then(result => {
  authStore.set({
    isAuthenticated: true,
    user: result.user,
    loading: false,
    error: null,
  });
});

// When session expires
authStore.set({
  isAuthenticated: false,
  user: null,
  loading: false,
  error: 'Session expired',
});
```

---

## 🎨 **Loading States**

### **Frontend Loading Pattern**:
```svelte
<script>
  import { authStore } from '$lib/stores/auth';
  
  let localLoading = false;
  
  async function handleAction() {
    localLoading = true;
    try {
      await auth.someAction();
    } finally {
      localLoading = false;
    }
  }
</script>

{#if localLoading || $authStore.loading}
  <LoadingSpinner />
{:else}
  <button on:click={handleAction}>
    Submit
  </button>
{/if}
```

---

## 📊 **Response Time Expectations**

Frontend should handle these backend response times:

| Endpoint | Expected | P95 | Timeout |
|----------|----------|-----|---------|
| POST /login | 100-200ms | 300ms | 5s |
| POST /register | 200-400ms | 600ms | 10s |
| POST /refresh | 50-100ms | 150ms | 3s |
| POST /logout | 50-100ms | 150ms | 3s |
| GET /me | 20-50ms | 100ms | 3s |
| OAuth2 flows | 500-1000ms | 1.5s | 15s |

### **Timeout Handling**:
```typescript
async function fetchWithTimeout(url: string, options: RequestInit, timeout = 5000) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);
  
  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
    });
    clearTimeout(timeoutId);
    return response;
  } catch (error) {
    clearTimeout(timeoutId);
    if (error.name === 'AbortError') {
      throw new Error('Request timed out. Please try again.');
    }
    throw error;
  }
}
```

---

## 🔐 **Security Contracts**

### **HTTPS Only (Production)**:
```typescript
// Enforce HTTPS in production
if (import.meta.env.PROD && !API_BASE_URL.startsWith('https://')) {
  console.error('API must use HTTPS in production');
}
```

### **Token Storage**:
```typescript
// SecureTokenStorage (localStorage)
// Future: Migrate to httpOnly cookies
class SecureTokenStorage {
  private static readonly KEY = 'bome_auth_tokens';
  
  static setTokens(tokens: Tokens): void {
    localStorage.setItem(this.KEY, JSON.stringify(tokens));
  }
  
  static getTokens(): Tokens | null {
    const stored = localStorage.getItem(this.KEY);
    return stored ? JSON.parse(stored) : null;
  }
  
  static clearTokens(): void {
    localStorage.removeItem(this.KEY);
  }
}
```

### **CORS Requirements**:
Backend must set these headers:
```
Access-Control-Allow-Origin: https://yourdomain.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
```

---

## 📝 **TypeScript Interfaces**

### **Shared Types**:
```typescript
// frontend/src/lib/types/auth.ts

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  email_verified: boolean;
  role: string;
  profile_picture_url: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login: string | null;
}

export interface LoginResponse {
  success: boolean;
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: User;
}

export interface RegisterResponse {
  success: boolean;
  message: string;
  user_id: string;
  email: string;
}

export interface TokenResponse {
  success: boolean;
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface ApiError {
  error: string;
  code: string;
  details?: Record<string, any>;
}
```

---

## 🎯 **Contract Validation**

### **Backend Must Provide**:
✅ All endpoints as documented  
✅ Correct HTTP status codes  
✅ Consistent error format  
✅ CORS headers  
✅ JWT tokens in correct format  
✅ Response times within SLAs  

### **Frontend Must Provide**:
✅ Correct request format  
✅ Authorization headers  
✅ Handle all error codes  
✅ Implement token refresh  
✅ Timeout handling  
✅ Loading states  

---

## 🔄 **Contract Evolution**

### **When Adding New Endpoints**:
1. ✅ Define endpoint in backend elastic band first
2. ✅ Implement backend handler
3. ✅ Add endpoint to `AUTH_ENDPOINTS` in frontend
4. ✅ Add method to auth store
5. ✅ Update this contract document
6. ✅ Add TypeScript interfaces if needed

### **When Modifying Responses**:
1. ⚠️ **Breaking change**: Requires version bump
2. ⚠️ Update both backend and frontend elastic bands
3. ⚠️ Update TypeScript interfaces
4. ⚠️ Test all consuming components
5. ⚠️ Document migration path

---

## 📍 **Related Files**

**Frontend Files**:
- Auth Store: `frontend/src/lib/auth.ts`
- API Client: `frontend/src/lib/api/client.ts`
- Types: `frontend/src/lib/types/auth.ts`
- Error Handler: `frontend/src/lib/api/errorHandler.ts`

**Backend Contract**:
- API Definitions: `_braids/authentication/backend/layers/application/ELASTIC-BAND-UP.md`
- Business Logic: `_braids/authentication/backend/layers/business-logic/ELASTIC-BAND-UP.md`

---

**Last Updated**: October 14, 2025  
**Contract Version**: 1.0  
**Status**: ✅ Production Stable  
**Breaking Changes**: Require major version bump

