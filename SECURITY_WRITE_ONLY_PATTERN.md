# 🔒 WRITE-ONLY SECURITY PATTERN

**Pattern**: Secret keys are **WRITE-ONLY** from the frontend  
**Last Updated**: October 28, 2025

---

## 🎯 **THE RULE**

```
┌─────────────┐                    ┌─────────────┐
│   FRONTEND  │                    │   BACKEND   │
│   (Admin)   │                    │  (Service)  │
└─────────────┘                    └─────────────┘
       │                                   │
       │  POST /settings/stripe-key       │
       │  { value: "sk_live_xxx" }        │
       ├──────────────────────────────────>│
       │                                   │ ✅ Store in DB
       │  { status: "success" }           │ secure_settings
       │<──────────────────────────────────┤
       │                                   │
       │  ❌ NO GET ENDPOINT EXISTS        │
       │  ❌ Value NEVER returned          │
       │                                   │
       │                                   │ ✅ Backend reads
       │                                   │ internally for
       │                                   │ Stripe API calls
```

---

## ✅ **WHAT FRONTEND CAN DO**

### **Admin Settings UI**

```typescript
// ✅ UPDATE secret (write-only)
async function updateStripeKey(newKey: string) {
    const response = await apiRequest('/admin/settings/secure', {
        method: 'POST',
        body: {
            key: 'stripe_secret_key',
            value: newKey
        }
    });
    
    if (response.status === 'success') {
        toast.success('Stripe key updated successfully');
        // Value is now stored, frontend never sees it again
    }
}
```

### **Admin Settings List (Metadata Only)**

```typescript
// ✅ GET list of configured settings (metadata only, no values)
async function getSecureSettings() {
    const response = await apiRequest('/admin/settings/secure/list');
    
    // Response:
    // {
    //   "status": "success",
    //   "data": [
    //     {
    //       "key": "stripe_secret_key",
    //       "is_configured": true,
    //       "updated_at": "2025-10-28T10:30:00Z"
    //       // ✅ NO "value" field!
    //     }
    //   ]
    // }
}
```

---

## ❌ **WHAT FRONTEND CANNOT DO**

```typescript
// ❌ WRONG - This endpoint should NOT exist
async function getStripeKey() {
    const response = await apiRequest('/admin/settings/stripe-key');
    return response.value;  // ❌ Secret should NEVER be returned
}

// ❌ WRONG - Displaying secret values
<input 
    type="text" 
    value={stripeSecretKey}  // ❌ Should never have this value
/>

// ✅ CORRECT - Input for UPDATING only
<input 
    type="password" 
    placeholder="Enter new Stripe secret key"
    value={newKeyInput}  // ✅ Only holds new input, not stored value
    onChange={(e) => setNewKeyInput(e.target.value)}
/>
<button onClick={() => updateStripeKey(newKeyInput)}>
    Update Key
</button>
```

---

## ✅ **WHAT BACKEND CAN DO**

### **Internal Use (Services)**

```go
// ✅ CORRECT - Backend reads for internal service use (with base64 decoding)
import "encoding/base64"

func initializeStripe(db *database.DB) error {
    var encodedKey string
    err := db.QueryRow(`
        SELECT value FROM secure_settings 
        WHERE key = 'stripe_secret_key'
    `).Scan(&encodedKey)
    
    if err != nil {
        return fmt.Errorf("failed to get Stripe key: %w", err)
    }
    
    // Decode base64-encoded key
    decodedBytes, err := base64.StdEncoding.DecodeString(encodedKey)
    if err != nil {
        return fmt.Errorf("failed to decode Stripe key: %w", err)
    }
    
    stripe.Key = string(decodedBytes)  // ✅ Used internally, never sent to frontend
    return nil
}
```

### **Admin Update Endpoint**

```go
// ✅ CORRECT - Accept updates, confirm without returning value
func UpdateSecureSetting(c *gin.Context) {
    // Super admin only
    if c.GetString("user_role") != "super_admin" {
        c.JSON(403, gin.H{"error": "Forbidden"})
        return
    }
    
    var req struct {
        Key   string `json:"key"`
        Value string `json:"value"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Store in database
    _, err := db.Exec(`
        INSERT INTO secure_settings (key, value, updated_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (key) DO UPDATE SET
            value = EXCLUDED.value,
            updated_at = NOW()
    `, req.Key, req.Value)
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to update"})
        return
    }
    
    // ✅ CORRECT - Confirm success WITHOUT returning value
    c.JSON(200, gin.H{
        "status": "success",
        "message": fmt.Sprintf("'%s' updated successfully", req.Key),
        // ❌ DO NOT INCLUDE: "value": req.Value
    })
}
```

### **Admin List Endpoint (Safe)**

```go
// ✅ CORRECT - Return metadata only
func GetSecureSettingsList(c *gin.Context) {
    rows, err := db.Query(`
        SELECT key, created_at, updated_at
        FROM secure_settings
        ORDER BY key
    `)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to get settings"})
        return
    }
    defer rows.Close()
    
    var settings []gin.H
    for rows.Next() {
        var key string
        var createdAt, updatedAt time.Time
        rows.Scan(&key, &createdAt, &updatedAt)
        
        settings = append(settings, gin.H{
            "key": key,
            "created_at": createdAt,
            "updated_at": updatedAt,
            "is_configured": true,  // ✅ Show it exists
            // ❌ NO "value" field
        })
    }
    
    c.JSON(200, gin.H{"status": "success", "data": settings})
}
```

---

## ❌ **WHAT BACKEND CANNOT DO**

```go
// ❌ WRONG - Never return secret values in API responses
func GetStripeSetting(c *gin.Context) {
    var key string
    db.QueryRow("SELECT value FROM secure_settings WHERE key = 'stripe_secret_key'").Scan(&key)
    
    c.JSON(200, gin.H{
        "stripe_key": key  // ❌ NEVER DO THIS!
    })
}

// ❌ WRONG - Never log secret values
log.Printf("Stripe key: %s", stripeKey)  // ❌ Secret in logs!

// ❌ WRONG - Never include secrets in error messages
return fmt.Errorf("failed to use key %s: %w", stripeKey, err)  // ❌ Secret in error!

// ✅ CORRECT - Generic error messages
log.Printf("Failed to load Stripe key from database")  // ✅ No value
return fmt.Errorf("failed to initialize Stripe: %w", err)  // ✅ No value
```

---

## 📋 **IMPLEMENTATION CHECKLIST**

When implementing secure settings:

### **Database**
- [ ] `secure_settings` table exists
- [ ] No FK relationships that could expose values
- [ ] Audit logging for updates (optional but recommended)

### **Backend**
- [ ] GET endpoint returns metadata only (key, updated_at, is_configured)
- [ ] POST/PUT endpoint for updates (super_admin only)
- [ ] Internal services read from DB for use
- [ ] No secret values in logs
- [ ] No secret values in error messages
- [ ] No secret values in API responses

### **Frontend**
- [ ] Input field for UPDATING (type="password")
- [ ] No display of current secret value
- [ ] Success message after update
- [ ] List shows "✓ Configured" instead of actual value
- [ ] No console.log of secret values

---

## 🎯 **UI EXAMPLES**

### **Admin Settings Page (Good)**

```svelte
<div class="setting-item">
    <h3>Stripe Secret Key</h3>
    
    <!-- ✅ CORRECT - Shows status, not value -->
    <p class="status">
        {#if stripeKeyConfigured}
            ✓ Configured (Last updated: {stripeKeyUpdatedAt})
        {:else}
            ⚠️ Not configured
        {/if}
    </p>
    
    <!-- ✅ CORRECT - Input for updating only -->
    <input 
        type="password" 
        bind:value={newStripeKey}
        placeholder="Enter new secret key"
    />
    
    <button on:click={updateStripeKey}>
        Update Key
    </button>
</div>
```

### **Admin Settings Page (Bad)**

```svelte
<!-- ❌ WRONG - Displaying secret value -->
<div class="setting-item">
    <h3>Stripe Secret Key</h3>
    <p>Current value: {currentStripeKey}</p>  <!-- ❌ Should NEVER exist -->
    
    <input 
        type="text"
        value={currentStripeKey}  <!-- ❌ Should NEVER exist -->
    />
</div>
```

---

## 🔍 **AUDIT QUESTIONS**

Ask yourself when implementing settings:

1. ❓ Can the frontend retrieve the secret value? → **Should be NO**
2. ❓ Is the secret value ever in an API response? → **Should be NO**
3. ❓ Can I see the secret in browser DevTools? → **Should be NO**
4. ❓ Can I see the secret in browser Network tab? → **Should be NO**
5. ❓ Can I see the secret in server logs? → **Should be NO**
6. ❓ Can super_admin update the secret? → **Should be YES**
7. ❓ Does backend use the secret internally? → **Should be YES**

---

## 🚨 **SECURITY INCIDENT: WHAT IF A SECRET IS EXPOSED?**

If a secret key is accidentally exposed:

1. **Immediately rotate the key**
   - Generate new key in Stripe/service
   - Update via admin UI
   - Old key is now invalid

2. **Check audit logs**
   - Who accessed what?
   - When was it exposed?
   - How was it exposed?

3. **Fix the vulnerability**
   - Remove GET endpoint if it exists
   - Remove value from API responses
   - Remove console.log statements
   - Remove value from error messages

4. **Verify fix**
   - Run through audit questions above
   - Check DevTools Network tab
   - Check server logs
   - Test with non-admin user

---

**Remember: If you can see a secret key in the browser, something is wrong!**

