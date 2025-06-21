# BOME API Documentation

Complete REST API reference for the Book of Mormon Evidences (BOME) streaming platform.

## 📋 Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Base URL & Versioning](#base-url--versioning)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [API Endpoints](#api-endpoints)
- [WebSocket API](#websocket-api)
- [SDK & Libraries](#sdk--libraries)
- [Examples](#examples)

## Overview

The BOME API is a RESTful web service that enables developers to integrate with the BOME streaming platform. It provides access to user management, video content, subscriptions, analytics, and administrative functions.

### Key Features
- **RESTful Design**: Standard HTTP methods and status codes
- **JSON Format**: All requests and responses use JSON
- **JWT Authentication**: Secure token-based authentication
- **Rate Limited**: Built-in rate limiting for stability
- **Versioned**: API versioning for backward compatibility
- **WebSocket Support**: Real-time updates and streaming

## Authentication

### JWT Token Authentication

All API requests require authentication using JWT tokens in the Authorization header:

```http
Authorization: Bearer <your-jwt-token>
```

### Getting a Token

**Endpoint**: `POST /api/v1/auth/login`

```json
{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response**:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 123,
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "role": "user"
  }
}
```

### Token Refresh

**Endpoint**: `POST /api/v1/auth/refresh`

Headers:
```http
Authorization: Bearer <your-current-token>
```

## Base URL & Versioning

- **Base URL**: `https://api.bome.com`
- **Current Version**: `v1`
- **Full Base URL**: `https://api.bome.com/api/v1`

All endpoints are prefixed with `/api/v1` for version 1 of the API.

## Response Format

### Success Response
```json
{
  "success": true,
  "data": {
    // Response data here
  },
  "message": "Operation completed successfully"
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  }
}
```

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid request data |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `AUTHENTICATION_ERROR` | Authentication failed |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `SERVER_ERROR` | Internal server error |

## Rate Limiting

- **Rate Limit**: 1000 requests per hour per API key
- **Burst Limit**: 100 requests per minute
- **Headers**: Rate limit information included in response headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## API Endpoints

### Authentication Endpoints

#### Login
```http
POST /api/v1/auth/login
```

**Request Body**:
```json
{
  "email": "string",
  "password": "string"
}
```

#### Register
```http
POST /api/v1/auth/register
```

**Request Body**:
```json
{
  "email": "string",
  "password": "string",
  "firstName": "string",
  "lastName": "string"
}
```

#### Logout
```http
POST /api/v1/auth/logout
```

#### Password Reset
```http
POST /api/v1/auth/password-reset
```

**Request Body**:
```json
{
  "email": "string"
}
```

### User Management Endpoints

#### Get Current User
```http
GET /api/v1/users/me
```

#### Update User Profile
```http
PUT /api/v1/users/profile
```

**Request Body**:
```json
{
  "firstName": "string",
  "lastName": "string",
  "bio": "string",
  "location": "string",
  "website": "string",
  "phone": "string"
}
```

#### Get User by ID
```http
GET /api/v1/users/{userId}
```

### Video Management Endpoints

#### Get Videos
```http
GET /api/v1/videos
```

**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)
- `category`: Filter by category
- `search`: Search query
- `sort`: Sort by (title, date, views, rating)
- `order`: Sort order (asc, desc)

#### Get Video by ID
```http
GET /api/v1/videos/{videoId}
```

#### Get Video Stream URL
```http
GET /api/v1/videos/{videoId}/stream
```

**Response**:
```json
{
  "success": true,
  "data": {
    "streamUrl": "https://stream.bome.com/video/123/playlist.m3u8",
    "thumbnailUrl": "https://cdn.bome.com/thumbnails/123.jpg",
    "duration": 3600,
    "quality": ["720p", "1080p", "4K"]
  }
}
```

#### Upload Video (Admin)
```http
POST /api/v1/videos
```

**Request Body** (multipart/form-data):
```
video: file
title: string
description: string
category: string
tags: string[]
```

### Subscription Endpoints

#### Get Current Subscription
```http
GET /api/v1/subscriptions/current
```

#### Get Available Plans
```http
GET /api/v1/subscriptions/plans
```

#### Create Subscription
```http
POST /api/v1/subscriptions
```

**Request Body**:
```json
{
  "planId": "string",
  "paymentMethodId": "string"
}
```

#### Update Subscription
```http
PUT /api/v1/subscriptions/{subscriptionId}
```

#### Cancel Subscription
```http
DELETE /api/v1/subscriptions/{subscriptionId}
```

### Analytics Endpoints

#### Get User Analytics
```http
GET /api/v1/analytics/users
```

#### Get Video Analytics
```http
GET /api/v1/analytics/videos
```

#### Get Revenue Analytics
```http
GET /api/v1/analytics/revenue
```

### Admin Endpoints

#### Get All Users (Admin)
```http
GET /api/v1/admin/users
```

#### Get System Stats (Admin)
```http
GET /api/v1/admin/stats
```

#### Moderate Content (Admin)
```http
POST /api/v1/admin/moderate
```

## WebSocket API

### Connection
```javascript
const ws = new WebSocket('wss://api.bome.com/ws');
```

### Authentication
```javascript
ws.send(JSON.stringify({
  type: 'auth',
  token: 'your-jwt-token'
}));
```

### Events

#### Video Progress
```javascript
// Send progress update
ws.send(JSON.stringify({
  type: 'video_progress',
  videoId: 123,
  progress: 45.5,
  timestamp: 1640995200
}));
```

#### Real-time Notifications
```javascript
// Receive notification
{
  type: 'notification',
  data: {
    id: 456,
    message: "New video available",
    timestamp: 1640995200
  }
}
```

## SDK & Libraries

### JavaScript/Node.js
```bash
npm install @bome/api-client
```

```javascript
import { BomeClient } from '@bome/api-client';

const client = new BomeClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.bome.com'
});

// Get videos
const videos = await client.videos.list();
```

### Python
```bash
pip install bome-python-sdk
```

```python
from bome import BomeClient

client = BomeClient(api_key='your-api-key')
videos = client.videos.list()
```

### PHP
```bash
composer require bome/php-sdk
```

```php
use Bome\Client;

$client = new Client('your-api-key');
$videos = $client->videos()->list();
```

## Examples

### Complete Authentication Flow

```javascript
// Login
const loginResponse = await fetch('https://api.bome.com/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
});

const { token } = await loginResponse.json();

// Use token for authenticated requests
const videosResponse = await fetch('https://api.bome.com/api/v1/videos', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const videos = await videosResponse.json();
```

### Video Streaming Integration

```javascript
// Get video stream URL
const streamResponse = await fetch(`https://api.bome.com/api/v1/videos/${videoId}/stream`, {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const { streamUrl } = await streamResponse.json();

// Use with video player
const video = document.getElementById('video-player');
video.src = streamUrl;
```

### Subscription Management

```javascript
// Get available plans
const plansResponse = await fetch('https://api.bome.com/api/v1/subscriptions/plans', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const plans = await plansResponse.json();

// Create subscription
const subscriptionResponse = await fetch('https://api.bome.com/api/v1/subscriptions', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    planId: 'premium-monthly',
    paymentMethodId: 'pm_1234567890'
  })
});
```

### Error Handling Best Practices

```javascript
async function makeApiRequest(url, options = {}) {
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
        ...options.headers
      }
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`API Error: ${error.error.message}`);
    }

    return await response.json();
  } catch (error) {
    console.error('API request failed:', error);
    throw error;
  }
}
```

## Testing

### Postman Collection
Download our [Postman collection](./postman-collection.json) for testing all endpoints.

### Test Environment
- **Base URL**: `https://api-staging.bome.com`
- **Test Credentials**: Contact support for test account access

## Support

For API support and questions:
- **Email**: api-support@bome.com
- **Documentation**: https://docs.bome.com/api
- **Status Page**: https://status.bome.com
- **GitHub Issues**: https://github.com/bome/api-issues

---

**Last Updated**: December 2024  
**API Version**: 1.0.0  
**Maintained By**: BOME API Team 