# 📊 Video Analytics Metrics Guide

## 🎯 **Complete Metrics Dictionary**

---

## 📈 **View Metrics**

### **`views`** (Total Views)
**Definition**: Total number of times a video was played  
**Calculation**: `COUNT(*)` from `video_views`  
**Use Case**: Overall popularity

```sql
SELECT COUNT(*) AS views
FROM video_views
WHERE video_id = $1;
```

**Good**: 1000+ views  
**Average**: 100-1000 views  
**Needs Improvement**: <100 views  

---

### **`unique_views`** (Unique Viewers)
**Definition**: Number of distinct users who watched  
**Calculation**: `COUNT(DISTINCT user_id)` from `video_views`  
**Use Case**: Actual audience size

```sql
SELECT COUNT(DISTINCT COALESCE(user_id, session_id)) AS unique_views
FROM video_views
WHERE video_id = $1;
```

**Key Insight**: `views / unique_views` = average rewatches per user  
- Ratio of 1.0 = everyone watched once  
- Ratio of 2.0 = average user watched twice (high engagement!)  

---

### **`watch_time`** (Total Watch Time)
**Definition**: Sum of all seconds watched by all users  
**Calculation**: `SUM(watched_duration)` from `video_views`  
**Use Case**: Total engagement, ad revenue potential

```sql
SELECT 
    SUM(watched_duration) AS total_watch_time_seconds,
    SUM(watched_duration) / 3600.0 AS total_watch_time_hours
FROM video_views
WHERE video_id = $1;
```

**Example**: 10,000 hours watched = enough for monetization!

---

### **`average_watch_time`** (Avg Duration Per View)
**Definition**: Average seconds watched per view  
**Calculation**: `AVG(watched_duration)` from `video_views`  
**Use Case**: Content quality indicator

```sql
SELECT 
    AVG(watched_duration) AS avg_watch_time,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY watched_duration) AS median_watch_time,
    (SELECT duration FROM master_video_list WHERE id = $1) AS video_duration,
    AVG(watched_duration)::FLOAT / 
        (SELECT duration FROM master_video_list WHERE id = $1)::FLOAT * 100 AS avg_percentage_watched
FROM video_views
WHERE video_id = $1;
```

**Benchmarks**:
- **Excellent**: >80% of video duration
- **Good**: 50-80%
- **Average**: 30-50%
- **Poor**: <30%

---

## 🎯 **Engagement Metrics**

### **`completion_rate`** (% Who Finished)
**Definition**: Percentage of viewers who watched ≥95% of video  
**Calculation**: Views with `watched_percentage >= 95` / total views  
**Use Case**: Content quality, "stickiness"

```sql
SELECT 
    COUNT(CASE WHEN watched_percentage >= 95 THEN 1 END)::FLOAT /
    COUNT(*)::FLOAT * 100 AS completion_rate
FROM video_views
WHERE video_id = $1;
```

**Benchmarks by Video Length**:
- **Short (<5 min)**: 40-60% is good
- **Medium (5-15 min)**: 30-40% is good
- **Long (15+ min)**: 20-30% is good

---

### **`bounce_rate`** (% Who Left Early)
**Definition**: Percentage who watched <10 seconds  
**Calculation**: Views with `watched_duration < 10` / total views  
**Use Case**: Content relevance, thumbnail/title accuracy

```sql
SELECT 
    COUNT(CASE WHEN watched_duration < 10 THEN 1 END)::FLOAT /
    COUNT(*)::FLOAT * 100 AS bounce_rate
FROM video_views
WHERE video_id = $1;
```

**Interpretation**:
- **<20%**: Excellent (thumbnail/title match content)
- **20-40%**: Good
- **40-60%**: Average (consider better thumbnail)
- **>60%**: Poor (misleading title or poor quality)

---

### **`engagement_score`** (Composite 0-100)
**Definition**: Weighted combination of all engagement metrics  
**Calculation**: Proprietary algorithm  
**Use Case**: Single number to compare videos

```sql
WITH metrics AS (
    SELECT 
        video_id,
        COUNT(*)::FLOAT AS views,
        COUNT(CASE WHEN watched_percentage >= 95 THEN 1 END)::FLOAT / 
            COUNT(*)::FLOAT * 100 AS completion_rate,
        AVG(watched_percentage) AS avg_watch_percentage,
        COUNT(CASE WHEN watched_duration < 10 THEN 1 END)::FLOAT / 
            COUNT(*)::FLOAT * 100 AS bounce_rate
    FROM video_views
    WHERE video_id = $1
    GROUP BY video_id
),
video_data AS (
    SELECT 
        v.id,
        v.likes::FLOAT / NULLIF(m.views, 0) * 100 AS likes_per_view
    FROM master_video_list v
    JOIN metrics m ON m.video_id = v.id
    WHERE v.id = $1
)
SELECT 
    (
        (m.completion_rate * 0.4) +                    -- 40% weight
        (m.avg_watch_percentage * 0.3) +               -- 30% weight
        ((100 - m.bounce_rate) * 0.2) +                -- 20% weight
        (LEAST(v.likes_per_view * 100, 10) * 0.1)      -- 10% weight (capped at 10)
    ) AS engagement_score
FROM metrics m
JOIN video_data v ON v.id = m.video_id;
```

**Score Ranges**:
- **90-100**: Viral content! 🔥
- **75-89**: Excellent, promote this!
- **60-74**: Good, solid performer
- **45-59**: Average
- **<45**: Needs improvement

---

## 👥 **User Behavior Metrics**

### **`first_view_dropout`** (When Do Users Leave?)
**Definition**: Distribution of when viewers stop watching  
**Use Case**: Identify boring parts of video

```sql
SELECT 
    FLOOR(watched_duration / 30) * 30 AS time_bucket, -- 30-second buckets
    COUNT(*) AS dropoffs
FROM video_views
WHERE video_id = $1 AND watched_percentage < 95
GROUP BY time_bucket
ORDER BY time_bucket;
```

**Visualization**: Heatmap showing drop-off points  
**Action**: Re-edit video to remove slow parts

---

### **`replay_rate`** (% Who Watched Multiple Times)
**Definition**: Users who watched same video >1 time  
**Use Case**: Educational content, rewatchability

```sql
WITH user_views AS (
    SELECT 
        user_id,
        COUNT(*) AS view_count
    FROM video_views
    WHERE video_id = $1 AND user_id IS NOT NULL
    GROUP BY user_id
)
SELECT 
    COUNT(CASE WHEN view_count > 1 THEN 1 END)::FLOAT /
    COUNT(*)::FLOAT * 100 AS replay_rate,
    AVG(CASE WHEN view_count > 1 THEN view_count END) AS avg_replays
FROM user_views;
```

**High Replay Rate** (>20%) = Educational/Reference content

---

### **`binge_watch_rate`** (% Who Watched Next Video)
**Definition**: Users who watched another video within 10 minutes  
**Use Case**: Content series, playlist effectiveness

```sql
WITH next_video AS (
    SELECT 
        v1.user_id,
        v1.created_at AS first_view,
        MIN(v2.created_at) AS next_view
    FROM video_views v1
    LEFT JOIN video_views v2 ON 
        v2.user_id = v1.user_id AND
        v2.video_id != v1.video_id AND
        v2.created_at > v1.created_at AND
        v2.created_at < v1.created_at + INTERVAL '10 minutes'
    WHERE v1.video_id = $1
    GROUP BY v1.user_id, v1.created_at
)
SELECT 
    COUNT(CASE WHEN next_view IS NOT NULL THEN 1 END)::FLOAT /
    COUNT(*)::FLOAT * 100 AS binge_watch_rate
FROM next_video;
```

---

## 💰 **Revenue Metrics**

### **`conversion_rate`** (% Who Subscribed After Watching)
**Definition**: Viewers who subscribed within 7 days  
**Use Case**: Revenue attribution

```sql
SELECT 
    COUNT(DISTINCT vv.user_id) AS total_viewers,
    COUNT(DISTINCT CASE 
        WHEN usc.created_at > vv.created_at 
         AND usc.created_at < vv.created_at + INTERVAL '7 days'
        THEN vv.user_id 
    END) AS converted_users,
    COUNT(DISTINCT CASE ... END)::FLOAT / 
    COUNT(DISTINCT vv.user_id)::FLOAT * 100 AS conversion_rate_7d
FROM video_views vv
LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = vv.user_id
WHERE vv.video_id = $1 AND vv.user_id IS NOT NULL;
```

**Top Conversion Videos** = Best for marketing!

---

### **`subscriber_views`** vs **`free_views`**
**Definition**: Views from paid vs free users  
**Use Case**: Content value perception

```sql
SELECT 
    COUNT(CASE WHEN usc.user_id IS NOT NULL THEN 1 END) AS subscriber_views,
    COUNT(CASE WHEN usc.user_id IS NULL THEN 1 END) AS free_views,
    COUNT(CASE WHEN usc.user_id IS NOT NULL THEN 1 END)::FLOAT /
    COUNT(*)::FLOAT * 100 AS subscriber_view_percentage
FROM video_views vv
LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = vv.user_id
WHERE vv.video_id = $1;
```

**High Subscriber %** = Premium content working!

---

### **`revenue_per_view`** (Estimated)
**Definition**: Revenue attributed to video views  
**Calculation**: (Subscribers from this video × Monthly price) / Total views

```sql
WITH conversions AS (
    SELECT COUNT(DISTINCT usc.user_id) AS new_subscribers
    FROM video_views vv
    JOIN user_stripe_customers_v2 usc ON usc.user_id = vv.user_id
    WHERE vv.video_id = $1
      AND usc.created_at > vv.created_at
      AND usc.created_at < vv.created_at + INTERVAL '7 days'
),
views AS (
    SELECT COUNT(*) AS total_views
    FROM video_views
    WHERE video_id = $1
)
SELECT 
    (c.new_subscribers * 9.97) / NULLIF(v.total_views, 0) AS revenue_per_view
FROM conversions c, views v;
```

---

## 📊 **Performance Metrics**

### **`views_per_day`** (Growth Trend)
**Definition**: Daily view count over time  
**Use Case**: Viral tracking, trend identification

```sql
SELECT 
    DATE(created_at) AS date,
    COUNT(*) AS views,
    COUNT(DISTINCT COALESCE(user_id, session_id)) AS unique_views
FROM video_views
WHERE video_id = $1
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

---

### **`trending_score`** (Hot Right Now)
**Definition**: Recent popularity with time decay  
**Use Case**: "Trending" section

```sql
WITH recent_stats AS (
    SELECT 
        video_id,
        COUNT(*) AS last_24h_views,
        MAX(created_at) AS last_view_at
    FROM video_views
    WHERE created_at > NOW() - INTERVAL '24 hours'
    GROUP BY video_id
),
video_stats AS (
    SELECT 
        v.id,
        v.title,
        COALESCE(r.last_24h_views, 0) AS recent_views,
        EXTRACT(EPOCH FROM (NOW() - COALESCE(r.last_view_at, v.created_at))) / 3600 AS hours_since_view,
        m.completion_rate,
        v.likes
    FROM master_video_list v
    LEFT JOIN recent_stats r ON r.video_id = v.id
    LEFT JOIN video_metrics m ON m.video_id = v.id AND m.date = CURRENT_DATE
)
SELECT 
    id,
    title,
    recent_views,
    (
        (recent_views / 24.0 * 0.5) +                                    -- Velocity (views/hour)
        ((completion_rate + likes * 10) / 2 * 0.3)                       -- Engagement
    ) * EXP(-hours_since_view / 72) AS trending_score                   -- Time decay (3 days)
FROM video_stats
WHERE recent_views > 0
ORDER BY trending_score DESC
LIMIT 10;
```

---

## 🌍 **Geographic Metrics**

### **`top_locations`** (Where Are Viewers?)
**Definition**: Geographic distribution of views  
**Use Case**: Content localization

```sql
-- Requires GeoIP lookup service
SELECT 
    country,
    COUNT(*) AS views,
    COUNT(DISTINCT user_id) AS unique_viewers
FROM video_views vv
JOIN geoip_lookup(vv.ip_address) geo ON true
WHERE vv.video_id = $1
GROUP BY country
ORDER BY views DESC;
```

---

## 🕒 **Temporal Metrics**

### **`peak_viewing_hours`** (When Do People Watch?)
**Definition**: Views by hour of day  
**Use Case**: Optimal upload time

```sql
SELECT 
    EXTRACT(HOUR FROM created_at) AS hour_of_day,
    COUNT(*) AS views
FROM video_views
WHERE video_id = $1
GROUP BY hour_of_day
ORDER BY hour_of_day;
```

---

## 🎯 **Content Quality Indicators**

### **Metric Interpretation Matrix**

| Metric | Excellent | Good | Average | Poor |
|--------|-----------|------|---------|------|
| **Completion Rate** | >50% | 30-50% | 15-30% | <15% |
| **Bounce Rate** | <20% | 20-40% | 40-60% | >60% |
| **Avg % Watched** | >75% | 50-75% | 30-50% | <30% |
| **Engagement Score** | >80 | 60-80 | 40-60 | <40 |
| **Replay Rate** | >20% | 10-20% | 5-10% | <5% |
| **Conversion Rate** | >5% | 2-5% | 1-2% | <1% |

---

## 🚀 **Actionable Insights**

### **High Bounce Rate** (>50%)
**Possible Causes**:
- Misleading thumbnail/title
- Poor video quality
- Slow start (boring intro)
- Wrong target audience

**Actions**:
- A/B test different thumbnails
- Cut first 10 seconds
- Add hook in first 5 seconds
- Review content targeting

---

### **Low Completion Rate** (<20%)
**Possible Causes**:
- Video too long
- Boring middle section
- Poor pacing
- Technical issues

**Actions**:
- Analyze drop-off points
- Edit out slow parts
- Add chapter markers
- Test different lengths

---

### **High Conversion Rate** (>3%)
**Success Indicators**:
- Great content
- Clear value proposition
- Effective CTA (call-to-action)

**Actions**:
- Promote this video!
- Use as marketing material
- Create similar content
- Link in email campaigns

---

## 📈 **Dashboard Layout Recommendations**

### **Video Detail Page:**
1. **Hero Metrics** (Large cards)
   - Views
   - Unique Viewers
   - Engagement Score
   - Revenue Generated

2. **Engagement Chart**
   - Line graph: Views over time
   - Heatmap: Drop-off points

3. **Demographics**
   - Subscriber vs Free
   - Top locations
   - Peak viewing hours

4. **Comparisons**
   - vs. Channel average
   - vs. Similar videos
   - Week-over-week growth

---

**Use these metrics to optimize your content strategy! 📊🚀**

