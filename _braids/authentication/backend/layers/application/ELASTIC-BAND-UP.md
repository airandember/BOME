# 🔗 ELASTIC BAND: Application → Presentation
**Interface Contract Between Backend API and Svelte5 Frontend**

---

## 📍 **Connection Points**

**From**: HTTP REST API (Layer 2 - Application)  
**To**: Svelte5 Frontend Components & Stores (Layer 1 - Presentation)  
**Purpose**: Define how frontend consumes backend API and manages auth state

---

## 🎯 **Frontend Auth Store Interface**

**File**: `frontend/src/lib/auth.ts`

### **Auth Store Structure**:
```typescript
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
}

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  email_verified: boolean;
  role: string;
  profile_picture_url: string | null;
}
```

### **Auth Store Methods**:
```typescript
// Registration
async register(data: {
  email: string;
  first_name: string;
  last_name: string;
}): Promise<RegisterResponse>

// Login
async login(email: string, password: string): Promise<LoginResponse>

// Logout
async logout(): Promise<void>

// Token refresh
async refreshToken(): Promise<TokenResponse>

// Get current user
async getCurrentUser(): Promise<User>

// Password reset
async forgotPassword(email: string): Promise<void>
async resetPassword(token: string, password: string): Promise<void>

// Email verification
async resendVerification(email: string, userId: string): Promise<void>
async setupPassword(token: string, userId: string, password: string): Promise<LoginResponse>
```

---

## 📡 **API Client Configuration**

### **Base Configuration**:
```typescript
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const API_VERSION = 'v1';
const API_URL = `${API_BASE_URL}/api/${API_VERSION}`;
```

### **Request Headers**:
```typescript
const headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
};

// For authenticated requests
if (accessToken) {
  headers['Authorization'] = `Bearer ${accessToken}`;
}
```

### **Request Interceptor**:
```typescript
// Automatically add auth token to requests
async function authenticatedFetch(url: string, options: RequestInit = {}) {
  const token = getAccessToken();
  
  const response = await fetch(url, {
    ...options,
    headers: {
      ...headers,
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
  });
  
  // Auto-refresh on 401
  if (response.status === 401 && token) {
    await refreshToken();
    // Retry request with new token
    return authenticatedFetch(url, options);
  }
  
  return response;
}
```

---

## 🔄 **Data Flow Patterns**

### **1. Login Flow**:
```typescript
// Component calls store method
const handleLogin = async () => {
  loading = true;
  error = '';
  
  try {
    const result = await auth.login(email, password);
    
    if (result.success) {
      // Store tokens
      SecureTokenStorage.setTokens({
        accessToken: result.access_token,
        refreshToken: result.refresh_token,
        expiresAt: Date.now() + (result.expires_in * 1000),
      });
      
      // Update store state
      auth.setUser(result.user);
      auth.setAuthenticated(true);
      
      // Navigate to dashboard
      goto('/dashboard');
    }
  } catch (err) {
    error = err.message;
  } finally {
    loading = false;
  }
};
```

### **2. Registration Flow**:
```typescript
const handleRegister = async () => {
  loading = true;
  
  try {
    const result = await auth.register({
      email,
      first_name: firstName,
      last_name: lastName,
    });
    
    if (result.success) {
      // Redirect to verification page
      goto(`/auth/verify-email?email=${email}&user_id=${result.user_id}`);
    }
  } catch (err) {
    error = err.message;
  } finally {
    loading = false;
  }
};
```

### **3. Auto-Refresh Pattern**:
```typescript
// Check token expiration on app load
onMount(async () => {
  const tokens = SecureTokenStorage.getTokens();
  
  if (tokens) {
    const expiresIn = tokens.expiresAt - Date.now();
    
    // Refresh if expiring in < 5 minutes
    if (expiresIn < 5 * 60 * 1000) {
      try {
        await auth.refreshToken();
      } catch (err) {
        // Refresh failed, logout
        auth.logout();
      }
    } else {
      // Token still valid, load user
      await auth.getCurrentUser();
    }
  }
});
```

---

## 🔐 **Token Management**

### **SecureTokenStorage Interface**:
```typescript
class SecureTokenStorage {
  // Store tokens securely
  static setTokens(tokens: {
    accessToken: string;
    refreshToken: string;
    expiresAt: number;
  }): void
  
  // Retrieve tokens
  static getTokens(): {
    accessToken: string;
    refreshToken: string;
    expiresAt: number;
  } | null
  
  // Clear tokens
  static clearTokens(): void
  
  // Check if token is expired
  static isTokenExpired(): boolean
}
```

### **Token Storage Implementation**:
```typescript
// Stored in localStorage (future: httpOnly cookies)
const TOKENS_KEY = 'bome_auth_tokens';

static setTokens(tokens) {
  localStorage.setItem(TOKENS_KEY, JSON.stringify(tokens));
}

static getTokens() {
  const stored = localStorage.getItem(TOKENS_KEY);
  return stored ? JSON.parse(stored) : null;
}

static clearTokens() {
  localStorage.removeItem(TOKENS_KEY);
}
```

---

## 🎨 **UI Component Patterns**

### **Protected Route Pattern**:
```svelte
<!-- +layout.svelte -->
<script>
  import { auth } from '$lib/auth';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  
  onMount(async () => {
    if (!$auth.isAuthenticated) {
      goto('/login');
    }
  });
</script>

{#if $auth.isAuthenticated}
  <slot />
{:else}
  <LoadingSpinner />
{/if}
```

### **Login Form Pattern**:
```svelte
<!-- routes/login/+page.svelte -->
<script lang="ts">
  import { auth } from '$lib/auth';
  
  let email = '';
  let password = '';
  let loading = false;
  let error = '';
  
  async function handleSubmit() {
    loading = true;
    error = '';
    
    try {
      await auth.login(email, password);
      // Success handled by store
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <input bind:value={email} type="email" required />
  <input bind:value={password} type="password" required />
  <button type="submit" disabled={loading}>
    {loading ? 'Logging in...' : 'Log In'}
  </button>
  {#if error}
    <p class="error">{error}</p>
  {/if}
</form>
```

### **Auth-Aware Navigation**:
```svelte
<!-- components/Navigation.svelte -->
<script>
  import { auth } from '$lib/auth';
</script>

<nav>
  {#if $auth.isAuthenticated}
    <a href="/dashboard">Dashboard</a>
    <a href="/account/profile">Profile</a>
    <button on:click={() => auth.logout()}>
      Log Out
    </button>
  {:else}
    <a href="/login">Log In</a>
    <a href="/register">Sign Up</a>
  {/if}
</nav>
```

---

## ⚠️ **Error Handling Patterns**

### **Display Error Messages**:
```typescript
function getErrorMessage(error: any): string {
  // API error response
  if (error.response?.data?.error) {
    return error.response.data.error;
  }
  
  // Network error
  if (error.message === 'Failed to fetch') {
    return 'Network error. Please check your connection.';
  }
  
  // Generic error
  return 'An error occurred. Please try again.';
}
```

### **Error Code Handling**:
```typescript
try {
  await auth.login(email, password);
} catch (err) {
  switch (err.code) {
    case 'INVALID_CREDENTIALS':
      error = 'Invalid email or password';
      break;
    case 'EMAIL_NOT_VERIFIED':
      error = 'Please verify your email first';
      showResendButton = true;
      break;
    case 'ACCOUNT_SUSPENDED':
      error = 'Your account has been suspended';
      break;
    case 'RATE_LIMIT_EXCEEDED':
      error = 'Too many attempts. Please try again later.';
      break;
    default:
      error = getErrorMessage(err);
  }
}
```

---

## 🔄 **State Synchronization**

### **Reactive Auth State**:
```typescript
// Svelte store
export const auth = writable<AuthState>({
  isAuthenticated: false,
  user: null,
  loading: false,
  error: null,
});

// Derived stores
export const isAuthenticated = derived(auth, $auth => $auth.isAuthenticated);
export const currentUser = derived(auth, $auth => $auth.user);
export const userRole = derived(auth, $auth => $auth.user?.role);
```

### **Component Reactivity**:
```svelte
<script>
  import { auth, isAuthenticated, currentUser } from '$lib/auth';
</script>

{#if $isAuthenticated}
  <p>Welcome, {$currentUser.first_name}!</p>
{/if}
```

---

## 📱 **Loading States**

### **Global Loading**:
```svelte
{#if $auth.loading}
  <div class="loading-overlay">
    <Spinner />
    <p>Loading...</p>
  </div>
{/if}
```

### **Component Loading**:
```svelte
<button type="submit" disabled={loading}>
  {#if loading}
    <Spinner size="small" />
    <span>Logging in...</span>
  {:else}
    Log In
  {/if}
</button>
```

---

## 🎯 **Form Validation**

### **Client-Side Validation**:
```typescript
// Email validation
function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

// Password strength
function isStrongPassword(password: string): boolean {
  return (
    password.length >= 8 &&
    /[A-Z]/.test(password) &&
    /[a-z]/.test(password) &&
    /[0-9]/.test(password)
  );
}
```

### **Real-Time Validation**:
```svelte
<script>
  $: emailValid = isValidEmail(email);
  $: passwordStrong = isStrongPassword(password);
</script>

<input
  bind:value={email}
  type="email"
  class:invalid={email && !emailValid}
/>
{#if email && !emailValid}
  <p class="validation-error">Please enter a valid email</p>
{/if}

<input
  bind:value={password}
  type="password"
  class:weak={password && !passwordStrong}
/>
{#if password && !passwordStrong}
  <ul class="password-requirements">
    <li class:met={password.length >= 8}>At least 8 characters</li>
    <li class:met={/[A-Z]/.test(password)}>One uppercase letter</li>
    <li class:met={/[a-z]/.test(password)}>One lowercase letter</li>
    <li class:met={/[0-9]/.test(password)}>One number</li>
  </ul>
{/if}
```

---

## 🔒 **Security Considerations**

### **XSS Prevention**:
```svelte
<!-- Svelte automatically escapes HTML -->
<p>{user.first_name}</p>  <!-- Safe -->

<!-- Use @html only for trusted content -->
{@html trustedContent}  <!-- Dangerous, avoid -->
```

### **CSRF Protection**:
```typescript
// JWT tokens provide CSRF protection
// No additional CSRF token needed for API calls
```

### **Secure Token Storage**:
```typescript
// Consider migrating to httpOnly cookies
// Current: localStorage (accessible to JS)
// Future: httpOnly cookies (not accessible to JS)
```

---

## 📊 **Analytics Integration**:
```typescript
// Track auth events
function trackAuthEvent(event: string, data?: any) {
  analytics.track(event, {
    ...data,
    timestamp: new Date().toISOString(),
  });
}

// Usage
auth.login(email, password).then(() => {
  trackAuthEvent('user_login', {
    method: 'email',
    user_id: user.id,
  });
});
```

---

## 🌐 **Internationalization**:
```typescript
// Error message translation
function getLocalizedError(code: string): string {
  const translations = {
    'INVALID_CREDENTIALS': t('auth.errors.invalid_credentials'),
    'EMAIL_NOT_VERIFIED': t('auth.errors.email_not_verified'),
    // ...
  };
  
  return translations[code] || t('auth.errors.generic');
}
```

---

**Last Updated**: October 14, 2025  
**Frontend Framework**: Svelte 5  
**Maintained By**: Frontend Team

