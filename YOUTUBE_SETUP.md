# YouTube Integration Setup Guide

This guide explains how to set up the YouTube PubSubHubbub webhook integration for the BOME platform.

## Overview

The YouTube integration automatically fetches new videos from the **Book of Mormon Evidence** YouTube channel (`@BookofMormonEvidence`, Channel ID: `UCHp1EBgpKytZt_-j72EZ83Q`) and makes them available on the frontend.

## Architecture

```
YouTube Channel → PubSubHubbub Hub → Your Webhook → Database → Frontend
```

## Backend Setup

### 1. Environment Variables

Add these to your `.env` file:

```env
# YouTube Webhook Configuration
YOUTUBE_WEBHOOK_URL=https://your-domain.com/webhook/youtube
YOUTUBE_VERIFY_TOKEN=your-random-verify-token-here
YOUTUBE_WEBHOOK_SECRET=your-optional-webhook-secret
```

### 2. Database Migration

Run the YouTube videos table migration:

```bash
# Apply the migration
mysql -u root -p your_database < backend/migrations/004_create_youtube_videos_table.sql
```

### 3. Webhook Endpoint

The webhook endpoint is automatically available at:
- **Verification**: `GET /webhook/youtube`
- **Notifications**: `POST /webhook/youtube`

### 4. API Endpoints

The following API endpoints are available:

- `GET /api/v1/youtube/videos` - Get all YouTube videos
- `GET /api/v1/youtube/videos/latest` - Get latest videos (cached)
- `GET /api/v1/youtube/status` - Get integration status
- `POST /api/v1/youtube/subscribe` - Manually subscribe to channel
- `POST /api/v1/youtube/unsubscribe` - Manually unsubscribe

## Frontend Setup

### 1. YouTube Page

The YouTube videos are displayed at `/youtube` with:
- Video grid layout
- Search functionality
- Direct links to YouTube
- Responsive design

### 2. Navigation Dropdown

The navigation now includes a "Videos" dropdown with:
- **YouTube** (Free, public access)
- **Premium** (Subscription required)

## PubSubHubbub Setup

### 1. Subscribe to Channel

You can subscribe manually via API:

```bash
curl -X POST https://your-domain.com/api/v1/youtube/subscribe
```

Or programmatically:

```javascript
const response = await fetch('/api/v1/youtube/subscribe', {
  method: 'POST',
  headers: { 'Authorization': 'Bearer your-jwt-token' }
});
```

### 2. Webhook URL Requirements

Your webhook URL must:
- Be publicly accessible (HTTPS required for production)
- Respond to GET requests for verification
- Accept POST requests for notifications
- Return proper HTTP status codes

### 3. Verification Process

When you subscribe, YouTube will:
1. Send a GET request with verification parameters
2. Expect your webhook to return the `hub.challenge` parameter
3. Confirm the subscription if verification succeeds

## Testing

### 1. Local Development

For local testing, use a tool like ngrok:

```bash
# Install ngrok
npm install -g ngrok

# Expose your local server
ngrok http 8080

# Use the ngrok URL as your YOUTUBE_WEBHOOK_URL
# Example: https://abc123.ngrok.io/webhook/youtube
```

### 2. Manual Testing

Test the webhook verification:

```bash
curl "https://your-domain.com/webhook/youtube?hub.mode=subscribe&hub.topic=https://www.youtube.com/xml/feeds/videos.xml?channel_id=UCHp1EBgpKytZt_-j72EZ83Q&hub.challenge=test123&hub.verify_token=your-verify-token"
```

### 3. Sample Data

The migration includes sample YouTube videos for testing. You can also trigger a manual fetch:

```bash
curl -X POST https://your-domain.com/api/v1/youtube/subscribe
```

## Production Deployment

### 1. Domain Setup

Ensure your domain:
- Has a valid SSL certificate
- Can receive HTTP requests on port 80/443
- Has proper DNS configuration

### 2. Environment Configuration

Set production environment variables:

```env
YOUTUBE_WEBHOOK_URL=https://api.yourdomain.com/webhook/youtube
YOUTUBE_VERIFY_TOKEN=secure-random-token-here
ENV=production
```

### 3. Subscription Management

Subscribe to the channel in production:

```bash
# Subscribe
curl -X POST https://api.yourdomain.com/api/v1/youtube/subscribe

# Check status
curl https://api.yourdomain.com/api/v1/youtube/status

# Unsubscribe if needed
curl -X POST https://api.yourdomain.com/api/v1/youtube/unsubscribe
```

## Monitoring

### 1. Logs

Monitor webhook activity in your application logs:
- Verification requests
- Video notifications
- Database updates
- Cache updates

### 2. Status Endpoint

Check integration status:

```bash
curl https://api.yourdomain.com/api/v1/youtube/status
```

Response:
```json
{
  "channel_id": "UCHp1EBgpKytZt_-j72EZ83Q",
  "channel_name": "@BookofMormonEvidence", 
  "total_videos": 25,
  "last_updated": "2024-01-15T10:30:00Z",
  "webhook_active": true
}
```

### 3. Cache Management

Videos are cached in `./cache/youtube_videos.json` for fast frontend loading.

## Troubleshooting

### Common Issues

1. **Webhook not receiving notifications**
   - Check webhook URL accessibility
   - Verify SSL certificate
   - Ensure proper HTTP status responses

2. **Verification failing**
   - Check verify token matches
   - Ensure GET endpoint works
   - Verify topic URL format

3. **Videos not updating**
   - Check database connection
   - Monitor webhook logs
   - Verify cache updates

### Debug Commands

```bash
# Test webhook verification
curl -v "https://your-domain.com/webhook/youtube?hub.mode=subscribe&hub.topic=https://www.youtube.com/xml/feeds/videos.xml?channel_id=UCHp1EBgpKytZt_-j72EZ83Q&hub.challenge=test&hub.verify_token=your-token"

# Check API endpoints
curl https://your-domain.com/api/v1/youtube/videos/latest

# Manual subscription
curl -X POST https://your-domain.com/api/v1/youtube/subscribe
```

## Security Considerations

1. **Verify Token**: Use a secure, random verify token
2. **HTTPS**: Always use HTTPS in production
3. **Rate Limiting**: Implement rate limiting on webhook endpoints
4. **Input Validation**: Validate all webhook data
5. **CORS**: Configure proper CORS headers

## Channel Information

- **Channel Name**: Book of Mormon Evidence
- **Channel Handle**: @BookofMormonEvidence
- **Channel ID**: UCHp1EBgpKytZt_-j72EZ83Q
- **Channel URL**: https://www.youtube.com/@BookofMormonEvidence
- **Feed URL**: https://www.youtube.com/xml/feeds/videos.xml?channel_id=UCHp1EBgpKytZt_-j72EZ83Q

## Support

For issues with the YouTube integration:
1. Check the application logs
2. Verify webhook configuration
3. Test API endpoints manually
4. Review PubSubHubbub documentation 