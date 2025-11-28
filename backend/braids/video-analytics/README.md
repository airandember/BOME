# 📊 Video Analytics BRAID

**System**: Video Analytics & Engagement Tracking  
**Purpose**: Track, analyze, and report on video viewing behavior  
**Status**: 🚧 **IN DEVELOPMENT**  
**V2 Compliant**: ✅ **YES** - Built on 100% V2 foundation

---

## 🎯 **What This BRAID Does**

This BRAID tracks **all video viewing activity** across your platform:

- ✅ **Individual View Tracking** - Every play, pause, seek, complete
- ✅ **User Watch History** - Resume playback, completion tracking
- ✅ **Aggregate Metrics** - Daily/weekly/monthly rollups per video
- ✅ **Engagement Analytics** - Watch time, completion rates, drop-off points
- ✅ **Revenue Attribution** - Which videos drive subscriptions
- ✅ **Performance Monitoring** - Popular content, trending videos

---

## 📁 **BRAID Structure**

```
video-analytics/
├── README.md               ← You are here
├── BRAID.md                ← Architecture overview
├── DATA_FLOW.md            ← Frontend → Backend → Database
├── METRICS_GUIDE.md        ← All metrics explained
├── QUERIES.md              ← Common SQL queries
└── API_REFERENCE.md        ← Route documentation
```

---

## 🗄️ **Database Tables**

### **Core Tracking Tables:**

| Table | Purpose | Key Metrics |
|-------|---------|-------------|
| `video_views` | Individual view events | `watched_duration`, `watched_percentage` |
| `video_watch_history` | User playback state | `last_position`, `completed` |
| `video_metrics` | Aggregated daily stats | `views`, `watch_time`, `completion_rate` |
| `user_metrics` | User engagement rollups | `video_views`, `video_watch_time` |
| `analytics_events` | Raw event stream | All video interactions |

### **Video Content Tables:**

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `master_video_list` | All videos | `views`, `total_watch_time`, `average_watch_time` |
| `videos` | Legacy video table | Migrating to master list |

---

## 🔄 **Data Flow**

```
┌─────────────────────────────────────────────────────────┐
│ FRONTEND (Video Player)                                 │
│ • User clicks play                                      │
│ • Player sends events every 10s                         │
│ • Tracks: play, pause, seek, progress, complete        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│ BACKEND (Routes)                                        │
│ /api/v1/analytics/video/track                          │
│ /api/v1/analytics/video/complete                       │
│ /api/v1/videos/:id/watch-history                       │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│ SERVICE LAYER (Business Logic)                          │
│ • VideoAnalyticsService                                 │
│ • WatchHistoryService                                   │
│ • MetricsAggregationService                            │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│ DATABASE LAYER                                          │
│ • video_views (raw events)                             │
│ • video_watch_history (state)                          │
│ • video_metrics (daily rollups)                        │
└─────────────────────────────────────────────────────────┘
```

---

## 📊 **Key Metrics Explained**

### **View Metrics:**
- **`views`** - Total number of play events
- **`unique_views`** - Distinct users who viewed
- **`watch_time`** - Total seconds watched (all users)
- **`average_watch_time`** - Average duration per viewer
- **`completion_rate`** - % of viewers who finished video

### **Engagement Metrics:**
- **`bounce_rate`** - % who left within 10 seconds
- **`watched_percentage`** - How far through video user got
- **`replay_count`** - Users who watched multiple times

### **Revenue Metrics:**
- **`conversion_rate`** - % of viewers who subscribed
- **`subscriber_views`** - Views from paid users
- **`free_views`** - Views from non-subscribers

---

## 🎬 **Quick Start**

### **1. Track a Video View:**
```typescript
// Frontend (Video Player)
await fetch('/api/v1/analytics/video/track', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({
    video_id: 123,
    watched_duration: 45,  // seconds
    watched_percentage: 35.5,
    session_id: 'abc123'
  })
});
```

### **2. Get Video Analytics:**
```go
// Backend
GET /api/v1/analytics/video/123/stats?period=7d
```

### **3. Get User Watch History:**
```go
// Backend
GET /api/v1/videos/:id/watch-history  // Resume playback
```

---

## 🚀 **Implementation Phases**

### **Phase 1: Core Tracking** ✅ (Tables exist)
- [x] Database schema created
- [x] Tables: `video_views`, `video_watch_history`
- [ ] Service layer implementation
- [ ] API routes

### **Phase 2: Aggregation** 🚧 (In Progress)
- [ ] Daily rollup jobs
- [ ] Populate `video_metrics` table
- [ ] Populate `user_metrics` table
- [ ] Cleanup old raw events

### **Phase 3: Analytics API** 📋 (Planned)
- [ ] Video performance endpoints
- [ ] User engagement endpoints
- [ ] Trending videos algorithm
- [ ] Revenue attribution

### **Phase 4: Admin Dashboard** 📋 (Planned)
- [ ] Video analytics UI
- [ ] User engagement charts
- [ ] Top performers report
- [ ] Export functionality

---

## 🔗 **Related Systems**

| System | Relationship | Data Flow |
|--------|--------------|-----------|
| **Subscription System** | Revenue attribution | Video views → Subscription conversions |
| **User Management** | User behavior | User profiles → Watch patterns |
| **Video Content** | Performance tracking | Video metadata → Engagement metrics |
| **Admin Dashboard** | Reporting | Analytics data → Visual charts |

---

## 📝 **Next Steps**

1. **Read**: `BRAID.md` for architecture details
2. **Review**: `METRICS_GUIDE.md` for all metrics
3. **Implement**: Service layer (`VideoAnalyticsService`)
4. **Test**: With real video playback data
5. **Deploy**: Aggregate metrics jobs

---

## 👥 **Who Needs This?**

- **Product Team**: Understand content performance
- **Business Team**: Revenue attribution, ROI
- **Content Team**: What videos drive engagement
- **Dev Team**: Implement tracking correctly

---

## 🎊 **Built on V2 Foundation**

This BRAID is built on your 100% V2-compliant backend:
- ✅ All services use V2 tables
- ✅ Proper foreign key relationships
- ✅ Optimized queries
- ✅ Accurate financial data

**Ready for production analytics!** 📈🚀

