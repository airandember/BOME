# DigitalOcean Deployment Configuration

## Environment Variables Setup

To fix the CORS issue where the frontend is trying to connect to `localhost:8080`, you need to set the following environment variables in your DigitalOcean App Platform deployment:

### Required Environment Variables

```bash
# Set this to your actual backend URL (replace with your actual backend app URL)
VITE_API_BASE_URL=https://your-backend-app.ondigitalocean.app/api/v1

# Optional: WebSocket URL for real-time features
VITE_WS_URL=wss://your-backend-app.ondigitalocean.app/ws

# App configuration
VITE_APP_NAME=Book of Mormon Evidences
VITE_APP_VERSION=1.0.0
```

### How to Set Environment Variables in DigitalOcean

1. **Go to your DigitalOcean App Platform dashboard**
2. **Select your frontend app**
3. **Go to Settings > Environment Variables**
4. **Add the following variables:**

| Variable Name | Value | Description |
|---------------|-------|-------------|
| `VITE_API_BASE_URL` | `https://your-backend-app.ondigitalocean.app/api/v1` | **CRITICAL**: Replace with your actual backend URL |
| `VITE_WS_URL` | `wss://your-backend-app.ondigitalocean.app/ws` | WebSocket URL (optional) |
| `VITE_APP_NAME` | `Book of Mormon Evidences` | App name |
| `VITE_APP_VERSION` | `1.0.0` | App version |

### Finding Your Backend URL

1. **In DigitalOcean App Platform dashboard**
2. **Look for your backend app component**
3. **Copy the public URL** (e.g., `https://bome-backend-abc123.ondigitalocean.app`)
4. **Add `/api/v1` to the end** for the `VITE_API_BASE_URL`

### Example Configuration

If your backend is deployed at `https://bome-backend-abc123.ondigitalocean.app`, then:

```bash
VITE_API_BASE_URL=https://bome-backend-abc123.ondigitalocean.app/api/v1
VITE_WS_URL=wss://bome-backend-abc123.ondigitalocean.app/ws
```

### After Setting Environment Variables

1. **Redeploy your frontend app**
2. **The CORS error should be resolved**
3. **Frontend will now connect to the deployed backend instead of localhost**

### Verification

After deployment, you can verify the configuration is working by:

1. **Opening your frontend app**
2. **Opening browser developer tools**
3. **Going to Network tab**
4. **Attempting to log in or make an API call**
5. **Verify the request goes to your backend URL, not localhost**

### Troubleshooting

If you still see CORS errors:

1. **Check that the environment variables are set correctly**
2. **Verify the backend URL is accessible**
3. **Ensure your backend CORS configuration allows your frontend domain**
4. **Redeploy the frontend after making changes**

### Backend CORS Configuration

Make sure your backend's CORS configuration includes your frontend domain:

```go
// In your backend CORS configuration
CORSAllowedOrigins: []string{
    "https://your-frontend-app.ondigitalocean.app",
    "http://localhost:5173", // for local development
    "http://localhost:4173", // for preview
}
```
