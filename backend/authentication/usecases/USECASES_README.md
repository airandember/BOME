# 🎯 Authentication Use Cases

This directory contains **pure business logic** for authentication operations, completely separated from HTTP concerns.

---

## 📋 **What Are Use Cases?**

Use cases represent **business operations** that can be triggered from any interface (HTTP, gRPC, CLI, tests). They contain:
- ✅ Business logic only (no HTTP/Gin code)
- ✅ Input validation
- ✅ Business rules enforcement  
- ✅ Orchestration of services and repositories
- ❌ NO HTTP status codes
- ❌ NO request/response parsing
- ❌ NO middleware concerns

---

## 📁 **Available Use Cases**

### **1. RegisterUser** (`register_user.go`)
**Purpose:** Handle user registration business logic

**Input:**
```go
type RegisterUserInput struct {
    Email     string
    FirstName string
    LastName  string
}
```

**Output:**
```go
type RegisterUserOutput struct {
    User    *authModels.User
    Message string
}
```

**Business Rules:**
- Validates email format
- Validates names
- Checks for duplicate email
- Creates user with unverified email
- Sends verification email
- Returns user object

**Usage:**
```go
useCase := usecases.NewRegisterUser(db, cryptoSvc, emailService)
output, err := useCase.Execute(usecases.RegisterUserInput{
    Email:     "user@example.com",
    FirstName: "John",
    LastName:  "Doe",
})
```

---

### **2. LoginUser** (`login_user.go`)
**Purpose:** Handle user login business logic

**Input:**
```go
type LoginUserInput struct {
    Email      string
    Password   string
    ClientIP   string
    DeviceInfo string
    UserAgent  string
}
```

**Output:**
```go
type LoginUserOutput struct {
    User         *authModels.User
    AccessToken  string
    RefreshToken string
    ExpiresIn    int
    TokenType    string
    Message      string
}
```

**Business Rules:**
- Validates email format
- Verifies user exists
- Checks password
- Validates account is active
- Requires email verification
- Checks session limits
- Generates JWT tokens
- Creates session record
- Updates last login
- Creates audit log

**Usage:**
```go
useCase := usecases.NewLoginUser(db, cryptoSvc)
output, err := useCase.Execute(usecases.LoginUserInput{
    Email:      "user@example.com",
    Password:   "SecurePass123!",
    ClientIP:   "192.168.1.1",
    DeviceInfo: "Chrome on Windows",
    UserAgent:  "Mozilla/5.0...",
})
```

---

### **3. VerifyEmail** (`verify_email.go`)
**Purpose:** Handle email verification business logic

**Input:**
```go
type VerifyEmailInput struct {
    Token  string
    UserID int // Optional
}
```

**Output:**
```go
type VerifyEmailOutput struct {
    User    *authModels.User
    Message string
}
```

**Business Rules:**
- Validates token
- Finds user by token
- Checks if already verified
- Marks email as verified
- Clears verification token
- Creates audit log

**Usage:**
```go
useCase := usecases.NewVerifyEmail(db, cryptoSvc)
output, err := useCase.Execute(usecases.VerifyEmailInput{
    Token: "verification-token-here",
})
```

---

## 🎯 **Benefits**

### **1. Testability** ✅
Use cases can be unit tested without starting an HTTP server:
```go
func TestRegisterUser(t *testing.T) {
    // Mock dependencies
    mockDB := &MockDB{}
    mockCrypto := &MockCrypto{}
    mockEmail := &MockEmail{}
    
    // Create use case
    useCase := NewRegisterUser(mockDB, mockCrypto, mockEmail)
    
    // Test business logic
    output, err := useCase.Execute(RegisterUserInput{
        Email: "test@example.com",
        // ...
    })
    
    // Assert business logic worked correctly
    assert.NoError(t, err)
    assert.NotNil(t, output.User)
}
```

### **2. Reusability** 🔄
Same use case can be called from:
- HTTP handlers (Gin)
- gRPC services
- CLI commands
- Background jobs
- WebSocket handlers
- GraphQL resolvers

### **3. Clear Separation** 🎭
**Before (Handler with Business Logic):**
```go
func RegisterHandler(db *database.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // HTTP parsing
        var req RegisterRequest
        c.ShouldBindJSON(&req)
        
        // Business logic mixed in!
        if err := validateEmail(req.Email); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        // More business logic...
        user := createUser(req)
        
        // HTTP response
        c.JSON(200, user)
    }
}
```

**After (Handler Uses Use Case):**
```go
func RegisterHandler(useCase *usecases.RegisterUser) gin.HandlerFunc {
    return func(c *gin.Context) {
        // HTTP parsing
        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        // Call pure business logic
        output, err := useCase.Execute(usecases.RegisterUserInput{
            Email:     req.Email,
            FirstName: req.FirstName,
            LastName:  req.LastName,
        })
        
        // HTTP response
        if err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, gin.H{
            "user":    output.User,
            "message": output.Message,
        })
    }
}
```

### **4. Single Responsibility** 📝
- **Use Case:** Pure business logic
- **Handler:** HTTP adapter (parse request, call use case, format response)
- **Service:** Technical operations (crypto, email, payment)
- **Repository/Model:** Data access

---

## 🔄 **Migration Strategy**

1. ✅ **Create Use Case** - Extract business logic
2. ⏳ **Update Handler** - Call use case instead of inline logic
3. ⏳ **Write Tests** - Test use case independently
4. ⏳ **Repeat** - For each handler

---

## 🚀 **Next Steps**

1. Update existing handlers to use these use cases
2. Write unit tests for each use case
3. Extract more use cases (RefreshToken, ChangePassword, etc.)
4. Apply pattern to other braids (Subscription, Video, etc.)

---

**Built with ❤️ for clean architecture!**

