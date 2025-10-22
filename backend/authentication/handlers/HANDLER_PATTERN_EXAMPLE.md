# 🎯 Handler Pattern: Before & After Use Cases

This document shows the **transformation** of handlers from thick to thin after extracting use cases.

---

## 📋 **The Pattern**

### **Before Use Cases** ❌
**Handler = HTTP + Business Logic (MIXED)**

```go
func RegisterHandler(db *database.DB, emailService *email.EmailService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // HTTP concerns
        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        // BUSINESS LOGIC MIXED IN! ❌
        req.Email = strings.ToLower(crypto.SanitizeString(req.Email))
        
        if err := crypto.ValidateEmail(req.Email); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        existingUser, _ := db.GetUserByEmail(req.Email)
        if existingUser != nil {
            c.JSON(http.StatusConflict, gin.H{"error": "User exists"})
            return
        }

        user := &User{
            Email:     req.Email,
            FirstName: req.FirstName,
            LastName:  req.LastName,
        }

        if err := db.CreateUser(user); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
            return
        }

        token := crypto.GenerateSecureToken()
        db.SetVerificationToken(user.ID, token)
        emailService.SendVerificationEmail(user.Email, token, user.FirstName)

        // HTTP response
        c.JSON(http.StatusOK, gin.H{"user": user})
    }
}
```

**Problems:**
- ❌ Can't test business logic without HTTP
- ❌ Can't reuse logic in CLI/gRPC/etc
- ❌ Hard to unit test
- ❌ Handler is 50+ lines
- ❌ Business rules hidden in HTTP code

---

### **After Use Cases** ✅
**Handler = Thin HTTP Adapter**

```go
func RegisterHandler(registerUC *usecases.RegisterUser) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Parse HTTP request
        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
            return
        }

        // 2. Call use case (pure business logic)
        output, err := registerUC.Execute(usecases.RegisterUserInput{
            Email:     req.Email,
            FirstName: req.FirstName,
            LastName:  req.LastName,
        })

        // 3. Handle errors with appropriate HTTP status
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // 4. Format HTTP response
        c.JSON(http.StatusOK, gin.H{
            "user":    output.User,
            "message": output.Message,
        })
    }
}
```

**Benefits:**
- ✅ Handler is only 20 lines
- ✅ Business logic is in use case (testable!)
- ✅ Can test use case without HTTP
- ✅ Can reuse use case in CLI/gRPC
- ✅ Clear separation of concerns

---

## 🔄 **Handler Transformation Guide**

### **Step 1: Identify Business Logic**
Look for code that:
- Validates data (not HTTP format)
- Enforces business rules
- Orchestrates services
- Makes decisions

### **Step 2: Extract to Use Case**
Move that logic to a use case:
```go
// Before: In handler
if len(password) < 12 {
    c.JSON(400, gin.H{"error": "Password too short"})
    return
}

// After: In use case
if len(input.Password) < 12 {
    return nil, fmt.Errorf("password must be at least 12 characters")
}
```

### **Step 3: Update Handler**
Make handler call use case:
```go
// Before
password := req.Password
if len(password) < 12 {
    c.JSON(400, gin.H{"error": "..."})
    return
}
// ... more logic

// After
output, err := useCase.Execute(input)
if err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}
```

---

## 🎯 **Real Example: LoginHandler**

### **Before** (100+ lines in handler)
```go
func LoginHandler(db *database.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req LoginRequest
        c.ShouldBindJSON(&req)
        
        // Validate email
        if err := crypto.ValidateEmail(req.Email); err != nil {
            c.JSON(400, gin.H{"error": "Invalid email"})
            return
        }
        
        // Get user
        user, err := db.GetUserByEmail(req.Email)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid credentials"})
            return
        }
        
        // Check password
        if err := crypto.CheckPassword(user.PasswordHash, req.Password); err != nil {
            c.JSON(401, gin.H{"error": "Invalid credentials"})
            return
        }
        
        // Check account status
        if !user.IsActive {
            c.JSON(403, gin.H{"error": "Account deactivated"})
            return
        }
        
        // Check email verified
        if !user.EmailVerified {
            c.JSON(403, gin.H{"error": "Email not verified"})
            return
        }
        
        // Check session limit
        canLogin, err := db.CheckSessionLimit(user.ID, user.MaxSessions)
        if err != nil {
            c.JSON(500, gin.H{"error": "Server error"})
            return
        }
        if !canLogin {
            c.JSON(429, gin.H{"error": "Too many sessions"})
            return
        }
        
        // Generate tokens
        tokens, err := crypto.GenerateTokenPair(user.ID, user.Email, user.Role, user.EmailVerified)
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to generate tokens"})
            return
        }
        
        // Create session
        session := &Session{...}
        db.CreateSession(session)
        
        // Update last login
        db.UpdateLastLogin(user.ID)
        
        // Create audit log
        db.CreateAuditLog(&AuditLog{...})
        
        // Return response
        c.JSON(200, gin.H{
            "access_token":  tokens.AccessToken,
            "refresh_token": tokens.RefreshToken,
            "user":          user,
        })
    }
}
```

### **After** (25 lines in handler)
```go
func LoginHandler(loginUC *usecases.LoginUser) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Parse request
        var req LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        // 2. Get client info for audit
        clientIP := crypto.GetGlobalCryptoService().GetClientIP(
            c.Request.RemoteAddr,
            c.GetHeader("X-Forwarded-For"),
            c.GetHeader("X-Real-IP"),
        )

        // 3. Execute use case
        output, err := loginUC.Execute(usecases.LoginUserInput{
            Email:      req.Email,
            Password:   req.Password,
            ClientIP:   clientIP,
            DeviceInfo: "Browser", // Could extract from user agent
            UserAgent:  c.Request.UserAgent(),
        })

        // 4. Handle result
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // 5. Return response
        c.JSON(http.StatusOK, gin.H{
            "access_token":  output.AccessToken,
            "refresh_token": output.RefreshToken,
            "expires_in":    output.ExpiresIn,
            "token_type":    output.TokenType,
            "user":          output.User,
        })
    }
}
```

**Difference:**
- **Before:** 100+ lines, all logic in handler
- **After:** 25 lines, logic in use case
- **Result:** 75% reduction in handler complexity!

---

## 🧪 **Testing Comparison**

### **Before: Testing Handler (Complex)**
```go
func TestRegisterHandler(t *testing.T) {
    // Need to set up entire HTTP context
    router := gin.Default()
    router.POST("/register", RegisterHandler(db, emailService))
    
    // Create HTTP request
    req := httptest.NewRequest("POST", "/register", body)
    w := httptest.NewRecorder()
    
    // Execute request
    router.ServeHTTP(w, req)
    
    // Parse HTTP response
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    // Assert
    assert.Equal(t, 200, w.Code)
}
```

### **After: Testing Use Case (Simple)**
```go
func TestRegisterUser(t *testing.T) {
    // Simple mocks
    mockDB := &MockDB{}
    mockCrypto := &MockCrypto{}
    mockEmail := &MockEmail{}
    
    // Test business logic directly
    useCase := NewRegisterUser(mockDB, mockCrypto, mockEmail)
    
    output, err := useCase.Execute(RegisterUserInput{
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
    })
    
    // Direct assertions
    assert.NoError(t, err)
    assert.NotNil(t, output.User)
    assert.Equal(t, "test@example.com", output.User.Email)
}
```

**Benefits:**
- ✅ No HTTP server needed
- ✅ No JSON parsing
- ✅ Direct business logic testing
- ✅ Faster tests
- ✅ Easier to write

---

## 🎯 **Migration Checklist**

For each handler:
- [ ] Identify business logic sections
- [ ] Create use case file
- [ ] Define input/output structs
- [ ] Move business logic to use case
- [ ] Update handler to call use case
- [ ] Write use case unit tests
- [ ] Verify handler still works

---

## 🚀 **Result**

**Before:**
- Handlers: 50-100+ lines
- Business logic: Mixed with HTTP
- Testing: Complex (needs HTTP)
- Reusability: Low

**After:**
- Handlers: 15-25 lines
- Business logic: In use cases
- Testing: Simple (no HTTP)
- Reusability: High

**Clean Architecture = Happy Developers!** 🎉


