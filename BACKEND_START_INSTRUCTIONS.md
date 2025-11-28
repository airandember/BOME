# Backend Start Instructions - CRITICAL

## ⚠️ ALWAYS Remember These Steps

### Correct Command Sequence
```powershell
# 1. Navigate to backend directory
cd S:\AirEmber\BOME\BOME\backend

# 2. (Optional) Kill old process if port is in use
$proc = netstat -ano | findstr ":8080" | findstr "LISTENING" | ForEach-Object { $_ -split '\s+' | Select-Object -Last 1 }
if ($proc) { Stop-Process -Id $proc -Force }

# 3. Start the backend
go run main.go
```

### Common Mistakes to AVOID ❌

1. **DON'T** run from `S:\AirEmber\BOME\BOME` (project root)
   - Will fail with: `CreateFile main.go: The system cannot find the file specified`

2. **DON'T** use relative paths that don't persist
   - PowerShell backgrounds lose directory context

3. **DON'T** use `cd backend` without full path
   - May not work reliably in background jobs

### Working Patterns ✅

**Pattern 1: Full Path (Most Reliable)**
```powershell
Set-Location -Path "S:\AirEmber\BOME\BOME\backend"; go run main.go
```

**Pattern 2: Inline cd with semicolon**
```powershell
cd S:\AirEmber\BOME\BOME\backend; go run main.go
```

**Pattern 3: If already in backend directory**
```powershell
# Just verify first:
pwd
# Should show: S:\AirEmber\BOME\BOME\backend
go run main.go
```

### Verification Steps

**Before running**:
```powershell
# Check current directory
Get-Location
# Should output: S:\AirEmber\BOME\BOME\backend

# Verify main.go exists
Test-Path .\main.go
# Should output: True
```

**After starting**:
```powershell
# Wait a few seconds then check logs
Start-Sleep -Seconds 5
Get-Content c:\Users\Owner\.cursor\projects\s-AirEmber-BOME-BOME\terminals\{TERMINAL_NUMBER}.txt | Select-Object -Last 5

# Should see:
# "🚀 Server starting on port 8080"
# NOT "CreateFile main.go: The system cannot find the file specified"
```

### Port 8080 Already in Use?

```powershell
# Find process using port 8080
netstat -ano | findstr ":8080" | findstr "LISTENING"

# Kill the process (replace PID with actual number)
Stop-Process -Id {PID} -Force

# Wait before restarting
Start-Sleep -Seconds 2
```

### Why This Matters

- ✅ `go run main.go` requires `main.go` to be in current directory
- ✅ PowerShell background jobs start in project root by default
- ✅ Directory context doesn't persist between command attempts
- ✅ Must use full paths or explicit `cd` before `go run`

## Quick Reference

| Current Directory | Command | Result |
|------------------|---------|--------|
| `S:\AirEmber\BOME\BOME` | `go run main.go` | ❌ File not found |
| `S:\AirEmber\BOME\BOME\backend` | `go run main.go` | ✅ Starts backend |
| Any directory | `cd S:\AirEmber\BOME\BOME\backend; go run main.go` | ✅ Starts backend |

---

**REMEMBER**: Always `cd S:\AirEmber\BOME\BOME\backend` first or use full path in the same command!

