# 📊 BOME Database Schema

**Database:** PostgreSQL 14+  
**Total Tables:** 74  
**Last Updated:** October 22, 2025  
**Status:** ✅ Production

---

## 📋 TABLE OF CONTENTS

1. [Authentication & Users](#authentication--users) (5 tables)
2. [Video Streaming](#video-streaming) (8 tables)
3. [Subscriptions & Billing](#subscriptions--billing) (15 tables)
4. [Creator Payouts](#creator-payouts) (5 tables) 
5. [Content Management](#content-management) (8 tables)
6. [Analytics](#analytics) (10 tables)
7. [Advertisements](#advertisements) (12 tables)
8. [Communication](#communication) (5 tables)
9. [System & Infrastructure](#system--infrastructure) (6 tables)

---

## 1. AUTHENTICATION & USERS

### `users`
**Purpose:** Core user accounts  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | User ID |
| email | VARCHAR(255) | UNIQUE, NOT NULL | Email address |
| password_hash | VARCHAR(255) | NOT NULL | Bcrypt hash |
| first_name | VARCHAR(100) | | First name |
| last_name | VARCHAR(100) | | Last name |
| role | VARCHAR(50) | DEFAULT 'user' | User role (user, super_admin, etc.) |
| email_verified | BOOLEAN | DEFAULT FALSE | Email verification status |
| created_at | TIMESTAMP | DEFAULT NOW() | Account creation |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update |
| last_login | TIMESTAMP | | Last login time |
| is_active | BOOLEAN | DEFAULT TRUE | Account active status |

**Indexes:**
- `idx_users_email` ON (email)
- `idx_users_role` ON (role)
- `idx_users_created_at` ON (created_at)

**Usage:** Authentication, authorization, user management

---

### `sessions`
**Purpose:** User session management  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Session ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User reference |
| token | VARCHAR(255) | UNIQUE, NOT NULL | Session token (JWT) |
| ip_address | INET | | Client IP |
| user_agent | TEXT | | Browser/client info |
| created_at | TIMESTAMP | DEFAULT NOW() | Session start |
| expires_at | TIMESTAMP | NOT NULL | Session expiration |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |

**Indexes:**
- `idx_sessions_user_id` ON (user_id)
- `idx_sessions_token` ON (token)
- `idx_sessions_expires_at` ON (expires_at)

**Usage:** Session tracking, logout, security auditing

---

### `oauth2_providers`
**Purpose:** OAuth2 provider configurations  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Provider ID |
| provider_name | VARCHAR(50) | UNIQUE, NOT NULL | google, facebook, etc. |
| client_id | VARCHAR(255) | NOT NULL | OAuth2 client ID |
| client_secret | VARCHAR(255) | NOT NULL | OAuth2 secret (encrypted) |
| redirect_url | VARCHAR(500) | NOT NULL | Callback URL |
| is_active | BOOLEAN | DEFAULT TRUE | Provider enabled |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Usage:** Google OAuth, social login

---

### `oauth2_accounts`
**Purpose:** Linked OAuth2 accounts  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Account link ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User reference |
| provider_id | INTEGER | FOREIGN KEY → oauth2_providers(id) | Provider |
| provider_user_id | VARCHAR(255) | NOT NULL | Provider's user ID |
| access_token | TEXT | | OAuth2 access token |
| refresh_token | TEXT | | OAuth2 refresh token |
| token_expires_at | TIMESTAMP | | Token expiration |
| profile_data | JSONB | | Provider profile JSON |
| linked_at | TIMESTAMP | DEFAULT NOW() | Link date |

**Indexes:**
- `idx_oauth2_accounts_user_id` ON (user_id)
- `idx_oauth2_accounts_provider_user` ON (provider_id, provider_user_id)

---

### `email_verification_tokens`
**Purpose:** Email verification tokens  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Token ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User reference |
| token | VARCHAR(255) | UNIQUE, NOT NULL | Verification token |
| created_at | TIMESTAMP | DEFAULT NOW() | Creation time |
| expires_at | TIMESTAMP | NOT NULL | Expiration |
| used_at | TIMESTAMP | | When verified |

**Usage:** Email verification flow

---

## 2. VIDEO STREAMING

### `master_video_list`
**Purpose:** Complete video catalog  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Video ID |
| bunny_video_id | VARCHAR(255) | UNIQUE | Bunny.net video ID |
| title | VARCHAR(500) | NOT NULL | Video title |
| description | TEXT | | Video description |
| duration | INTEGER | | Duration in seconds |
| thumbnail_url | VARCHAR(1000) | | Thumbnail image URL |
| stream_url | VARCHAR(1000) | | Streaming URL |
| status | VARCHAR(50) | DEFAULT 'pending' | encoding, ready, failed |
| is_public | BOOLEAN | DEFAULT FALSE | Public visibility |
| requires_subscription | BOOLEAN | DEFAULT TRUE | Subscription required |
| view_count | INTEGER | DEFAULT 0 | Total views |
| created_at | TIMESTAMP | DEFAULT NOW() | Upload date |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update |
| published_at | TIMESTAMP | | Publish date |

**Indexes:**
- `idx_master_video_list_bunny_id` ON (bunny_video_id)
- `idx_master_video_list_status` ON (status)
- `idx_master_video_list_published` ON (published_at)

**Usage:** Main video database, video management

---

### `video_tags`
**Purpose:** Video categorization tags  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Tag ID |
| video_id | INTEGER | FOREIGN KEY → master_video_list(id) | Video reference |
| tag_name | VARCHAR(100) | NOT NULL | Tag name |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_video_tags_video_id` ON (video_id)
- `idx_video_tags_tag_name` ON (tag_name)

---

### `video_metadata`
**Purpose:** Additional video metadata  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Metadata ID |
| video_id | INTEGER | UNIQUE, FOREIGN KEY → master_video_list(id) | Video reference |
| file_size | BIGINT | | File size in bytes |
| resolution | VARCHAR(20) | | 1080p, 720p, etc. |
| bitrate | INTEGER | | Bitrate in kbps |
| codec | VARCHAR(50) | | Video codec |
| encoding_progress | INTEGER | DEFAULT 0 | 0-100 percentage |
| seo_title | VARCHAR(500) | | SEO-optimized title |
| seo_description | TEXT | | SEO description |
| seo_keywords | TEXT | | SEO keywords |

---

### `youtube_videos`
**Purpose:** YouTube RSS feed videos  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Record ID |
| youtube_id | VARCHAR(255) | UNIQUE, NOT NULL | YouTube video ID |
| title | VARCHAR(500) | NOT NULL | Video title |
| description | TEXT | | Description |
| published_at | TIMESTAMP | NOT NULL | YouTube publish date |
| thumbnail_url | VARCHAR(1000) | | Thumbnail URL |
| channel_id | VARCHAR(255) | | YouTube channel ID |
| duration | INTEGER | | Duration in seconds |
| fetched_at | TIMESTAMP | DEFAULT NOW() | When we fetched it |

**Indexes:**
- `idx_youtube_videos_youtube_id` ON (youtube_id)
- `idx_youtube_videos_published` ON (published_at)

**Usage:** YouTube RSS integration

---

### `video_views`
**Purpose:** Video view tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | View ID |
| video_id | INTEGER | FOREIGN KEY → master_video_list(id) | Video reference |
| user_id | INTEGER | FOREIGN KEY → users(id) | User who viewed (nullable) |
| session_id | VARCHAR(255) | | Anonymous session ID |
| ip_address | INET | | Viewer IP |
| watched_duration | INTEGER | DEFAULT 0 | Seconds watched |
| watched_percentage | DECIMAL(5,2) | DEFAULT 0 | % completed |
| created_at | TIMESTAMP | DEFAULT NOW() | View timestamp |

**Indexes:**
- `idx_video_views_video_id` ON (video_id)
- `idx_video_views_user_id` ON (user_id)
- `idx_video_views_created_at` ON (created_at)

**Usage:** Analytics, engagement tracking

---

### `video_watch_history`
**Purpose:** User watch history  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | History ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User reference |
| video_id | INTEGER | FOREIGN KEY → master_video_list(id) | Video reference |
| last_position | INTEGER | DEFAULT 0 | Last playback position (seconds) |
| completed | BOOLEAN | DEFAULT FALSE | Finished watching |
| first_watched_at | TIMESTAMP | DEFAULT NOW() | First view |
| last_watched_at | TIMESTAMP | DEFAULT NOW() | Most recent view |

**Indexes:**
- `idx_video_watch_history_user_video` ON (user_id, video_id)

---

### `video_playlists`
**Purpose:** User-created playlists  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Playlist ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | Creator |
| name | VARCHAR(255) | NOT NULL | Playlist name |
| description | TEXT | | Description |
| is_public | BOOLEAN | DEFAULT FALSE | Public visibility |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

---

### `playlist_videos`
**Purpose:** Videos in playlists  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Entry ID |
| playlist_id | INTEGER | FOREIGN KEY → video_playlists(id) | Playlist |
| video_id | INTEGER | FOREIGN KEY → master_video_list(id) | Video |
| position | INTEGER | DEFAULT 0 | Order in playlist |
| added_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_playlist_videos_playlist_id` ON (playlist_id)

---

## 3. SUBSCRIPTIONS & BILLING

### `stripe_customers`
**Purpose:** Stripe customer records  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Record ID |
| user_id | INTEGER | UNIQUE, FOREIGN KEY → users(id) | User reference |
| stripe_customer_id | VARCHAR(255) | UNIQUE, NOT NULL | Stripe ID |
| email | VARCHAR(255) | | Customer email |
| name | VARCHAR(255) | | Customer name |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_stripe_customers_stripe_id` ON (stripe_customer_id)
- `idx_stripe_customers_user_id` ON (user_id)

---

### `subscription_plans`
**Purpose:** Available subscription tiers  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Plan ID |
| name | VARCHAR(255) | NOT NULL | Plan name (Basic, Premium, etc.) |
| description | TEXT | | Plan description |
| stripe_price_id | VARCHAR(255) | UNIQUE | Stripe price ID |
| stripe_product_id | VARCHAR(255) | | Stripe product ID |
| price | DECIMAL(10,2) | NOT NULL | Monthly price |
| billing_period | VARCHAR(50) | DEFAULT 'month' | month, year |
| features | JSONB | | Feature list JSON |
| is_active | BOOLEAN | DEFAULT TRUE | Available for signup |
| display_order | INTEGER | DEFAULT 0 | Sort order |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `subscriptions`
**Purpose:** Active user subscriptions  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Subscription ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | Subscriber |
| plan_id | INTEGER | FOREIGN KEY → subscription_plans(id) | Plan |
| stripe_subscription_id | VARCHAR(255) | UNIQUE | Stripe subscription ID |
| status | VARCHAR(50) | NOT NULL | active, canceled, past_due, etc. |
| current_period_start | TIMESTAMP | | Billing period start |
| current_period_end | TIMESTAMP | | Billing period end |
| cancel_at_period_end | BOOLEAN | DEFAULT FALSE | Scheduled cancellation |
| canceled_at | TIMESTAMP | | Cancellation date |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_subscriptions_user_id` ON (user_id)
- `idx_subscriptions_stripe_id` ON (stripe_subscription_id)
- `idx_subscriptions_status` ON (status)

---

### `subscription_history`
**Purpose:** Subscription change log  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | History ID |
| subscription_id | INTEGER | FOREIGN KEY → subscriptions(id) | Subscription |
| old_status | VARCHAR(50) | | Previous status |
| new_status | VARCHAR(50) | NOT NULL | New status |
| change_reason | TEXT | | Reason for change |
| changed_at | TIMESTAMP | DEFAULT NOW() | |

---

### `stripe_invoices`
**Purpose:** Stripe invoice records  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Invoice ID |
| stripe_invoice_id | VARCHAR(255) | UNIQUE, NOT NULL | Stripe invoice ID |
| customer_id | INTEGER | FOREIGN KEY → stripe_customers(id) | Customer |
| subscription_id | INTEGER | FOREIGN KEY → subscriptions(id) | Subscription |
| amount | DECIMAL(10,2) | NOT NULL | Invoice amount |
| currency | VARCHAR(10) | DEFAULT 'usd' | Currency code |
| status | VARCHAR(50) | NOT NULL | paid, open, void, etc. |
| invoice_pdf | VARCHAR(1000) | | PDF URL |
| created_at | TIMESTAMP | NOT NULL | Invoice date |

---

### `stripe_payments`
**Purpose:** Payment transaction records  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Payment ID |
| stripe_payment_intent_id | VARCHAR(255) | UNIQUE, NOT NULL | Stripe payment intent ID |
| customer_id | INTEGER | FOREIGN KEY → stripe_customers(id) | Customer |
| amount | DECIMAL(10,2) | NOT NULL | Payment amount |
| currency | VARCHAR(10) | DEFAULT 'usd' | Currency |
| status | VARCHAR(50) | NOT NULL | succeeded, failed, etc. |
| payment_method | VARCHAR(50) | | card, bank, etc. |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `stripe_refunds`
**Purpose:** Refund records  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Refund ID |
| stripe_refund_id | VARCHAR(255) | UNIQUE, NOT NULL | Stripe refund ID |
| payment_id | INTEGER | FOREIGN KEY → stripe_payments(id) | Original payment |
| amount | DECIMAL(10,2) | NOT NULL | Refund amount |
| reason | VARCHAR(255) | | Refund reason |
| status | VARCHAR(50) | NOT NULL | succeeded, failed, etc. |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `stripe_webhook_events`
**Purpose:** Webhook event log  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Event ID |
| stripe_event_id | VARCHAR(255) | UNIQUE, NOT NULL | Stripe event ID |
| event_type | VARCHAR(255) | NOT NULL | customer.created, etc. |
| payload | JSONB | NOT NULL | Full event JSON |
| processed | BOOLEAN | DEFAULT FALSE | Processing status |
| processed_at | TIMESTAMP | | Processing time |
| error_message | TEXT | | Error if failed |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_stripe_webhook_events_type` ON (event_type)
- `idx_stripe_webhook_events_processed` ON (processed)

---

### `subscription_offers`
**Purpose:** Promotional offers/coupons  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Offer ID |
| name | VARCHAR(255) | NOT NULL | Offer name |
| code | VARCHAR(100) | UNIQUE, NOT NULL | Promo code |
| stripe_coupon_id | VARCHAR(255) | UNIQUE | Stripe coupon ID |
| discount_type | VARCHAR(20) | NOT NULL | percent, amount |
| discount_value | DECIMAL(10,2) | NOT NULL | Discount value |
| duration | VARCHAR(20) | DEFAULT 'once' | once, repeating, forever |
| duration_months | INTEGER | | Months if repeating |
| max_redemptions | INTEGER | | Max uses |
| redemptions_count | INTEGER | DEFAULT 0 | Current uses |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| valid_from | TIMESTAMP | NOT NULL | Valid start |
| valid_until | TIMESTAMP | | Valid end |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `stripe_monthly_metrics`
**Purpose:** Monthly financial metrics  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Metric ID |
| month | DATE | UNIQUE, NOT NULL | Month (YYYY-MM-01) |
| mrr | DECIMAL(10,2) | DEFAULT 0 | Monthly Recurring Revenue |
| new_customers | INTEGER | DEFAULT 0 | New customers |
| churned_customers | INTEGER | DEFAULT 0 | Lost customers |
| total_active_subscriptions | INTEGER | DEFAULT 0 | Active subs |
| total_revenue | DECIMAL(10,2) | DEFAULT 0 | Total revenue |
| refund_amount | DECIMAL(10,2) | DEFAULT 0 | Total refunds |
| calculated_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_stripe_monthly_metrics_month` ON (month)

---

### `ghost_customers`
**Purpose:** Data inconsistency tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Ghost ID |
| type | VARCHAR(50) | NOT NULL | stripe_only, local_only, mismatch |
| stripe_customer_id | VARCHAR(255) | | Stripe ID |
| local_user_id | INTEGER | | Local user ID |
| issue_description | TEXT | | What's wrong |
| discovered_at | TIMESTAMP | DEFAULT NOW() | |
| resolved_at | TIMESTAMP | | Resolution time |
| resolution_notes | TEXT | | How it was fixed |

---

### `stripe_products`
**Purpose:** Stripe product catalog  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Product ID |
| stripe_product_id | VARCHAR(255) | UNIQUE, NOT NULL | Stripe product ID |
| name | VARCHAR(255) | NOT NULL | Product name |
| description | TEXT | | Description |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| metadata | JSONB | | Additional data |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `stripe_prices`
**Purpose:** Stripe price points  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Price ID |
| stripe_price_id | VARCHAR(255) | UNIQUE, NOT NULL | Stripe price ID |
| product_id | INTEGER | FOREIGN KEY → stripe_products(id) | Product |
| unit_amount | INTEGER | NOT NULL | Amount in cents |
| currency | VARCHAR(10) | DEFAULT 'usd' | Currency |
| recurring_interval | VARCHAR(20) | | month, year |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `subscriber_enhanced_stats`
**Purpose:** Enhanced subscriber analytics  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Stat ID |
| user_id | INTEGER | UNIQUE, FOREIGN KEY → users(id) | User |
| total_watch_time | INTEGER | DEFAULT 0 | Total seconds watched |
| videos_watched | INTEGER | DEFAULT 0 | Video count |
| last_activity_at | TIMESTAMP | | Last activity |
| engagement_score | DECIMAL(5,2) | DEFAULT 0 | 0-100 score |
| calculated_at | TIMESTAMP | DEFAULT NOW() | |

---

## 4. CREATOR PAYOUTS

### `presenters`
**Purpose:** Content creator registry  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Presenter ID |
| full_name | VARCHAR(255) | NOT NULL | Full name |
| email | VARCHAR(255) | UNIQUE, NOT NULL | Contact email |
| payment_email | VARCHAR(255) | | PayPal/payment email |
| bio | TEXT | | Biography |
| profile_image_url | VARCHAR(1000) | | Profile image |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| is_verified | BOOLEAN | DEFAULT FALSE | Verification status |
| total_videos | INTEGER | DEFAULT 0 | Video count |
| total_views | INTEGER | DEFAULT 0 | Total views |
| total_payouts | DECIMAL(10,2) | DEFAULT 0 | Lifetime earnings |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_presenters_email` ON (email)
- `idx_presenters_is_active` ON (is_active)

**Usage:** Creator management, payout distribution

---

### `video_presenters`
**Purpose:** Video-to-presenter linking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Link ID |
| video_id | INTEGER | FOREIGN KEY → master_video_list(id) | Video |
| presenter_id | INTEGER | FOREIGN KEY → presenters(id) | Presenter |
| role | VARCHAR(100) | DEFAULT 'presenter' | presenter, host, guest |
| percentage_share | DECIMAL(5,2) | DEFAULT 100.00 | Payout % (0-100) |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_video_presenters_video_id` ON (video_id)
- `idx_video_presenters_presenter_id` ON (presenter_id)

**Usage:** Revenue sharing, multi-presenter videos

---

### `payout_formulas`
**Purpose:** Payout calculation methods  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Formula ID |
| name | VARCHAR(255) | UNIQUE, NOT NULL | Formula name |
| description | TEXT | | How it works |
| formula_type | VARCHAR(50) | NOT NULL | flat, per_view, engagement, custom |
| base_rate | DECIMAL(10,2) | DEFAULT 0 | Base payment |
| per_view_rate | DECIMAL(10,4) | DEFAULT 0 | Rate per view |
| engagement_multiplier | DECIMAL(5,2) | DEFAULT 1.00 | Engagement factor |
| custom_formula | TEXT | | Custom SQL/logic |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| is_default | BOOLEAN | DEFAULT FALSE | Default formula |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Usage:** Flexible payout calculations

---

### `presenter_payouts`
**Purpose:** Monthly payout records  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Payout ID |
| presenter_id | INTEGER | FOREIGN KEY → presenters(id) | Presenter |
| formula_id | INTEGER | FOREIGN KEY → payout_formulas(id) | Formula used |
| month | DATE | NOT NULL | Payout month (YYYY-MM-01) |
| total_views | INTEGER | DEFAULT 0 | Views in month |
| total_watch_time | INTEGER | DEFAULT 0 | Watch time (seconds) |
| engagement_score | DECIMAL(5,2) | DEFAULT 0 | Engagement metric |
| calculated_amount | DECIMAL(10,2) | NOT NULL | Calculated payout |
| final_amount | DECIMAL(10,2) | | Final payout (after adjustments) |
| status | VARCHAR(50) | DEFAULT 'pending' | pending, approved, paid |
| approved_at | TIMESTAMP | | Approval time |
| paid_at | TIMESTAMP | | Payment time |
| notes | TEXT | | Admin notes |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_presenter_payouts_presenter_id` ON (presenter_id)
- `idx_presenter_payouts_month` ON (month)
- `idx_presenter_payouts_status` ON (status)

---

### `payout_transactions`
**Purpose:** Payment transaction log  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Transaction ID |
| payout_id | INTEGER | FOREIGN KEY → presenter_payouts(id) | Payout reference |
| presenter_id | INTEGER | FOREIGN KEY → presenters(id) | Presenter |
| amount | DECIMAL(10,2) | NOT NULL | Transaction amount |
| payment_method | VARCHAR(50) | NOT NULL | paypal, bank_transfer, etc. |
| transaction_id | VARCHAR(255) | UNIQUE | External transaction ID |
| status | VARCHAR(50) | DEFAULT 'pending' | pending, completed, failed |
| error_message | TEXT | | Error if failed |
| processed_at | TIMESTAMP | | Processing time |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_payout_transactions_payout_id` ON (payout_id)
- `idx_payout_transactions_status` ON (status)

---

## 5. CONTENT MANAGEMENT

### `tags`
**Purpose:** Tag taxonomy  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Tag ID |
| name | VARCHAR(100) | UNIQUE, NOT NULL | Tag name |
| slug | VARCHAR(100) | UNIQUE, NOT NULL | URL-friendly slug |
| description | TEXT | | Tag description |
| usage_count | INTEGER | DEFAULT 0 | Times used |
| is_featured | BOOLEAN | DEFAULT FALSE | Featured tag |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_tags_slug` ON (slug)
- `idx_tags_usage_count` ON (usage_count DESC)

---

### `categories`
**Purpose:** Content categories  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Category ID |
| name | VARCHAR(100) | UNIQUE, NOT NULL | Category name |
| slug | VARCHAR(100) | UNIQUE, NOT NULL | URL slug |
| parent_id | INTEGER | FOREIGN KEY → categories(id) | Parent category |
| description | TEXT | | Description |
| icon | VARCHAR(100) | | Icon name |
| display_order | INTEGER | DEFAULT 0 | Sort order |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `articles`
**Purpose:** Blog/news articles  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Article ID |
| title | VARCHAR(500) | NOT NULL | Article title |
| slug | VARCHAR(500) | UNIQUE, NOT NULL | URL slug |
| content | TEXT | NOT NULL | Article content (HTML/Markdown) |
| excerpt | TEXT | | Short summary |
| author_id | INTEGER | FOREIGN KEY → users(id) | Author |
| category_id | INTEGER | FOREIGN KEY → categories(id) | Category |
| featured_image | VARCHAR(1000) | | Header image |
| status | VARCHAR(50) | DEFAULT 'draft' | draft, published, archived |
| view_count | INTEGER | DEFAULT 0 | Views |
| published_at | TIMESTAMP | | Publish date |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

---

### `article_tags`
**Purpose:** Article-tag relationships  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Link ID |
| article_id | INTEGER | FOREIGN KEY → articles(id) | Article |
| tag_id | INTEGER | FOREIGN KEY → tags(id) | Tag |

---

### `comments`
**Purpose:** User comments  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Comment ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | Commenter |
| content_type | VARCHAR(50) | NOT NULL | video, article |
| content_id | INTEGER | NOT NULL | Video/Article ID |
| parent_id | INTEGER | FOREIGN KEY → comments(id) | Reply to comment |
| content | TEXT | NOT NULL | Comment text |
| is_approved | BOOLEAN | DEFAULT TRUE | Moderation |
| is_edited | BOOLEAN | DEFAULT FALSE | Edit flag |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

---

### `content_reports`
**Purpose:** User-reported content  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Report ID |
| reporter_id | INTEGER | FOREIGN KEY → users(id) | Who reported |
| content_type | VARCHAR(50) | NOT NULL | video, comment, article |
| content_id | INTEGER | NOT NULL | Content ID |
| reason | VARCHAR(255) | NOT NULL | Report reason |
| description | TEXT | | Details |
| status | VARCHAR(50) | DEFAULT 'pending' | pending, reviewed, actioned |
| reviewed_by | INTEGER | FOREIGN KEY → users(id) | Reviewer |
| reviewed_at | TIMESTAMP | | Review time |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `page_content`
**Purpose:** Static page content  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Page ID |
| slug | VARCHAR(255) | UNIQUE, NOT NULL | URL slug |
| title | VARCHAR(500) | NOT NULL | Page title |
| content | TEXT | NOT NULL | Page content |
| is_published | BOOLEAN | DEFAULT TRUE | Visibility |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

---

### `seo_metadata`
**Purpose:** SEO metadata for pages  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | SEO ID |
| content_type | VARCHAR(50) | NOT NULL | video, article, page |
| content_id | INTEGER | NOT NULL | Content ID |
| meta_title | VARCHAR(500) | | SEO title |
| meta_description | TEXT | | SEO description |
| meta_keywords | TEXT | | SEO keywords |
| og_image | VARCHAR(1000) | | Open Graph image |
| canonical_url | VARCHAR(1000) | | Canonical URL |

---

## 6. ANALYTICS

### `user_activity_log`
**Purpose:** User activity tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Log ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User (nullable) |
| session_id | VARCHAR(255) | | Session identifier |
| activity_type | VARCHAR(100) | NOT NULL | login, view_video, subscribe |
| activity_data | JSONB | | Additional data |
| ip_address | INET | | User IP |
| user_agent | TEXT | | Browser info |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_user_activity_log_user_id` ON (user_id)
- `idx_user_activity_log_activity_type` ON (activity_type)
- `idx_user_activity_log_created_at` ON (created_at)

---

### `video_analytics`
**Purpose:** Video performance metrics  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Metric ID |
| video_id | INTEGER | FOREIGN KEY → master_video_list(id) | Video |
| date | DATE | NOT NULL | Metric date |
| views | INTEGER | DEFAULT 0 | Daily views |
| unique_viewers | INTEGER | DEFAULT 0 | Unique viewers |
| total_watch_time | INTEGER | DEFAULT 0 | Total seconds |
| avg_watch_time | INTEGER | DEFAULT 0 | Avg seconds |
| completion_rate | DECIMAL(5,2) | DEFAULT 0 | % completed |
| engagement_score | DECIMAL(5,2) | DEFAULT 0 | Engagement metric |

**Indexes:**
- `idx_video_analytics_video_date` ON (video_id, date)

---

### `subscriber_metrics`
**Purpose:** Daily subscriber metrics  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Metric ID |
| date | DATE | UNIQUE, NOT NULL | Metric date |
| new_subscribers | INTEGER | DEFAULT 0 | New subs |
| canceled_subscribers | INTEGER | DEFAULT 0 | Canceled subs |
| total_active | INTEGER | DEFAULT 0 | Active count |
| mrr | DECIMAL(10,2) | DEFAULT 0 | Monthly Recurring Revenue |
| churn_rate | DECIMAL(5,2) | DEFAULT 0 | Churn % |

---

### `revenue_analytics`
**Purpose:** Revenue tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Revenue ID |
| date | DATE | NOT NULL | Revenue date |
| revenue_type | VARCHAR(50) | NOT NULL | subscription, ad, other |
| amount | DECIMAL(10,2) | NOT NULL | Amount |
| source | VARCHAR(100) | | Source identifier |
| currency | VARCHAR(10) | DEFAULT 'usd' | Currency |

---

### `system_metrics`
**Purpose:** System performance metrics  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Metric ID |
| timestamp | TIMESTAMP | DEFAULT NOW() | Metric time |
| metric_name | VARCHAR(100) | NOT NULL | cpu_usage, memory, etc. |
| metric_value | DECIMAL(10,2) | NOT NULL | Value |
| unit | VARCHAR(50) | | percentage, bytes, etc. |

---

### `search_analytics`
**Purpose:** Search query tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Search ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User (nullable) |
| query | VARCHAR(500) | NOT NULL | Search query |
| results_count | INTEGER | DEFAULT 0 | Results returned |
| clicked_result_id | INTEGER | | Which result clicked |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `engagement_metrics`
**Purpose:** User engagement tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Metric ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User |
| date | DATE | NOT NULL | Metric date |
| sessions | INTEGER | DEFAULT 0 | Session count |
| page_views | INTEGER | DEFAULT 0 | Pages viewed |
| time_on_site | INTEGER | DEFAULT 0 | Seconds on site |
| videos_watched | INTEGER | DEFAULT 0 | Videos watched |

---

### `conversion_events`
**Purpose:** Conversion tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Event ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User (nullable) |
| event_type | VARCHAR(100) | NOT NULL | signup, subscribe, etc. |
| source | VARCHAR(100) | | Traffic source |
| campaign | VARCHAR(100) | | Campaign ID |
| value | DECIMAL(10,2) | | Conversion value |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ab_tests`
**Purpose:** A/B test configuration  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Test ID |
| name | VARCHAR(255) | NOT NULL | Test name |
| description | TEXT | | Description |
| variant_a | JSONB | | Variant A config |
| variant_b | JSONB | | Variant B config |
| is_active | BOOLEAN | DEFAULT TRUE | Test status |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ab_test_assignments`
**Purpose:** User A/B test assignments  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Assignment ID |
| test_id | INTEGER | FOREIGN KEY → ab_tests(id) | Test |
| user_id | INTEGER | FOREIGN KEY → users(id) | User (nullable) |
| variant | VARCHAR(10) | NOT NULL | a, b |
| assigned_at | TIMESTAMP | DEFAULT NOW() | |

---

## 7. ADVERTISEMENTS

### `advertisers`
**Purpose:** Advertiser accounts  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Advertiser ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User account |
| company_name | VARCHAR(255) | NOT NULL | Company name |
| business_email | VARCHAR(255) | UNIQUE, NOT NULL | Business email |
| contact_name | VARCHAR(255) | | Contact person |
| contact_phone | VARCHAR(50) | | Phone number |
| industry | VARCHAR(100) | | Industry type |
| website | VARCHAR(500) | | Company website |
| status | VARCHAR(50) | DEFAULT 'pending' | pending, approved, rejected |
| approved_at | TIMESTAMP | | Approval time |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ad_campaigns`
**Purpose:** Advertising campaigns  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Campaign ID |
| advertiser_id | INTEGER | FOREIGN KEY → advertisers(id) | Advertiser |
| name | VARCHAR(255) | NOT NULL | Campaign name |
| description | TEXT | | Description |
| budget | DECIMAL(10,2) | NOT NULL | Total budget |
| daily_budget | DECIMAL(10,2) | | Daily spend limit |
| spent_amount | DECIMAL(10,2) | DEFAULT 0 | Amount spent |
| status | VARCHAR(50) | DEFAULT 'draft' | draft, active, paused, completed |
| start_date | TIMESTAMP | NOT NULL | Campaign start |
| end_date | TIMESTAMP | | Campaign end |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `advertisements`
**Purpose:** Individual ads  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Ad ID |
| campaign_id | INTEGER | FOREIGN KEY → ad_campaigns(id) | Campaign |
| title | VARCHAR(255) | NOT NULL | Ad title |
| description | TEXT | | Ad description |
| ad_type | VARCHAR(50) | NOT NULL | banner, video, native |
| image_url | VARCHAR(1000) | | Ad image |
| video_url | VARCHAR(1000) | | Ad video |
| click_url | VARCHAR(1000) | NOT NULL | Destination URL |
| cta_text | VARCHAR(100) | | Call-to-action text |
| status | VARCHAR(50) | DEFAULT 'pending' | pending, approved, active |
| impressions | INTEGER | DEFAULT 0 | Times shown |
| clicks | INTEGER | DEFAULT 0 | Click count |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ad_placements`
**Purpose:** Ad placement configuration  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Placement ID |
| name | VARCHAR(255) | UNIQUE, NOT NULL | Placement name |
| location | VARCHAR(100) | NOT NULL | homepage, video_page, sidebar |
| size | VARCHAR(50) | | 300x250, 728x90, etc. |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| priority | INTEGER | DEFAULT 0 | Display priority |

---

### `ad_impressions`
**Purpose:** Ad impression tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Impression ID |
| ad_id | INTEGER | FOREIGN KEY → advertisements(id) | Ad |
| user_id | INTEGER | FOREIGN KEY → users(id) | User (nullable) |
| placement_id | INTEGER | FOREIGN KEY → ad_placements(id) | Placement |
| ip_address | INET | | Viewer IP |
| user_agent | TEXT | | Browser info |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_ad_impressions_ad_id` ON (ad_id)
- `idx_ad_impressions_created_at` ON (created_at)

---

### `ad_clicks`
**Purpose:** Ad click tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Click ID |
| ad_id | INTEGER | FOREIGN KEY → advertisements(id) | Ad |
| impression_id | INTEGER | FOREIGN KEY → ad_impressions(id) | Related impression |
| user_id | INTEGER | FOREIGN KEY → users(id) | User (nullable) |
| ip_address | INET | | Clicker IP |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ad_analytics`
**Purpose:** Campaign performance  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Analytics ID |
| campaign_id | INTEGER | FOREIGN KEY → ad_campaigns(id) | Campaign |
| date | DATE | NOT NULL | Analytics date |
| impressions | INTEGER | DEFAULT 0 | Daily impressions |
| clicks | INTEGER | DEFAULT 0 | Daily clicks |
| ctr | DECIMAL(5,2) | DEFAULT 0 | Click-through rate % |
| cost | DECIMAL(10,2) | DEFAULT 0 | Daily cost |
| conversions | INTEGER | DEFAULT 0 | Conversions |

---

### `ad_targeting_rules`
**Purpose:** Ad targeting configuration  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Rule ID |
| campaign_id | INTEGER | FOREIGN KEY → ad_campaigns(id) | Campaign |
| target_type | VARCHAR(50) | NOT NULL | age, location, interest |
| target_value | JSONB | NOT NULL | Targeting criteria |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ad_billing`
**Purpose:** Advertiser billing records  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Billing ID |
| advertiser_id | INTEGER | FOREIGN KEY → advertisers(id) | Advertiser |
| campaign_id | INTEGER | FOREIGN KEY → ad_campaigns(id) | Campaign |
| amount | DECIMAL(10,2) | NOT NULL | Billing amount |
| billing_period_start | DATE | NOT NULL | Period start |
| billing_period_end | DATE | NOT NULL | Period end |
| status | VARCHAR(50) | DEFAULT 'pending' | pending, paid, overdue |
| invoice_number | VARCHAR(100) | UNIQUE | Invoice # |
| paid_at | TIMESTAMP | | Payment time |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ad_conversions`
**Purpose:** Conversion tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Conversion ID |
| ad_id | INTEGER | FOREIGN KEY → advertisements(id) | Ad |
| click_id | INTEGER | FOREIGN KEY → ad_clicks(id) | Click |
| user_id | INTEGER | FOREIGN KEY → users(id) | User (nullable) |
| conversion_type | VARCHAR(50) | NOT NULL | signup, purchase, etc. |
| conversion_value | DECIMAL(10,2) | | Conversion value |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ad_frauddetection`
**Purpose:** Fraud detection log  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Detection ID |
| ad_id | INTEGER | FOREIGN KEY → advertisements(id) | Ad |
| ip_address | INET | | Suspicious IP |
| fraud_type | VARCHAR(50) | NOT NULL | click_fraud, impression_fraud |
| confidence_score | DECIMAL(5,2) | DEFAULT 0 | 0-100 confidence |
| action_taken | VARCHAR(50) | | blocked, flagged, etc. |
| detected_at | TIMESTAMP | DEFAULT NOW() | |

---

### `ad_reports`
**Purpose:** Campaign reports  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Report ID |
| campaign_id | INTEGER | FOREIGN KEY → ad_campaigns(id) | Campaign |
| report_type | VARCHAR(50) | NOT NULL | daily, weekly, monthly |
| report_data | JSONB | NOT NULL | Report JSON |
| generated_at | TIMESTAMP | DEFAULT NOW() | |

---

## 8. COMMUNICATION

### `email_templates`
**Purpose:** Email template library  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Template ID |
| name | VARCHAR(255) | UNIQUE, NOT NULL | Template name |
| subject | VARCHAR(500) | NOT NULL | Email subject |
| html_body | TEXT | NOT NULL | HTML template |
| text_body | TEXT | | Plain text version |
| variables | JSONB | | Template variables |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### `email_log`
**Purpose:** Email sending log  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Log ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | Recipient |
| template_id | INTEGER | FOREIGN KEY → email_templates(id) | Template used |
| to_email | VARCHAR(255) | NOT NULL | Recipient email |
| subject | VARCHAR(500) | NOT NULL | Email subject |
| status | VARCHAR(50) | DEFAULT 'sent' | sent, failed, bounced |
| error_message | TEXT | | Error if failed |
| opened_at | TIMESTAMP | | Email opened |
| clicked_at | TIMESTAMP | | Link clicked |
| sent_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_email_log_user_id` ON (user_id)
- `idx_email_log_sent_at` ON (sent_at)

---

### `notifications`
**Purpose:** In-app notifications  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Notification ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | Recipient |
| title | VARCHAR(255) | NOT NULL | Notification title |
| message | TEXT | NOT NULL | Message content |
| type | VARCHAR(50) | DEFAULT 'info' | info, warning, success, error |
| action_url | VARCHAR(1000) | | Action link |
| is_read | BOOLEAN | DEFAULT FALSE | Read status |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_notifications_user_id` ON (user_id)
- `idx_notifications_is_read` ON (is_read)

---

### `notification_preferences`
**Purpose:** User notification settings  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Preference ID |
| user_id | INTEGER | UNIQUE, FOREIGN KEY → users(id) | User |
| email_new_video | BOOLEAN | DEFAULT TRUE | New video emails |
| email_marketing | BOOLEAN | DEFAULT TRUE | Marketing emails |
| email_subscription | BOOLEAN | DEFAULT TRUE | Subscription emails |
| in_app_notifications | BOOLEAN | DEFAULT TRUE | In-app notifications |
| push_notifications | BOOLEAN | DEFAULT FALSE | Push notifications |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

---

### `contact_messages`
**Purpose:** Contact form submissions  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Message ID |
| name | VARCHAR(255) | NOT NULL | Sender name |
| email | VARCHAR(255) | NOT NULL | Sender email |
| subject | VARCHAR(500) | | Message subject |
| message | TEXT | NOT NULL | Message content |
| status | VARCHAR(50) | DEFAULT 'new' | new, replied, closed |
| replied_at | TIMESTAMP | | Reply time |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

## 9. SYSTEM & INFRASTRUCTURE

### `migrations`
**Purpose:** Database migration tracking  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Migration ID |
| version | VARCHAR(50) | UNIQUE, NOT NULL | Migration version |
| name | VARCHAR(255) | NOT NULL | Migration name |
| executed_at | TIMESTAMP | DEFAULT NOW() | Execution time |

---

### `system_settings`
**Purpose:** Application configuration  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Setting ID |
| key | VARCHAR(255) | UNIQUE, NOT NULL | Setting key |
| value | TEXT | | Setting value |
| data_type | VARCHAR(50) | DEFAULT 'string' | string, int, bool, json |
| description | TEXT | | Setting description |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

---

### `audit_log`
**Purpose:** System audit trail  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Log ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | User |
| action | VARCHAR(255) | NOT NULL | Action performed |
| entity_type | VARCHAR(100) | | Affected entity type |
| entity_id | INTEGER | | Affected entity ID |
| old_values | JSONB | | Before state |
| new_values | JSONB | | After state |
| ip_address | INET | | User IP |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_audit_log_user_id` ON (user_id)
- `idx_audit_log_created_at` ON (created_at)

---

### `feature_flags`
**Purpose:** Feature toggle system  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Flag ID |
| name | VARCHAR(255) | UNIQUE, NOT NULL | Flag name |
| description | TEXT | | Description |
| is_enabled | BOOLEAN | DEFAULT FALSE | Flag status |
| rollout_percentage | INTEGER | DEFAULT 0 | % rollout (0-100) |
| target_users | JSONB | | Specific user targets |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

---

### `api_keys`
**Purpose:** API key management  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Key ID |
| user_id | INTEGER | FOREIGN KEY → users(id) | Owner |
| key_hash | VARCHAR(255) | UNIQUE, NOT NULL | Hashed API key |
| name | VARCHAR(255) | | Key name/description |
| permissions | JSONB | | Permission scopes |
| is_active | BOOLEAN | DEFAULT TRUE | Active status |
| last_used_at | TIMESTAMP | | Last use time |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| expires_at | TIMESTAMP | | Expiration |

---

### `rate_limits`
**Purpose:** API rate limiting  
**Primary Key:** `id` (SERIAL)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Limit ID |
| identifier | VARCHAR(255) | NOT NULL | IP or API key |
| endpoint | VARCHAR(255) | NOT NULL | API endpoint |
| request_count | INTEGER | DEFAULT 0 | Requests in window |
| window_start | TIMESTAMP | NOT NULL | Window start time |
| created_at | TIMESTAMP | DEFAULT NOW() | |

**Indexes:**
- `idx_rate_limits_identifier_endpoint` ON (identifier, endpoint)

---

## 📊 USAGE NOTES

### Foreign Key Relationships
- **CASCADE** delete for dependent records (e.g., sessions when user deleted)
- **SET NULL** for optional relationships (e.g., articles when author deleted)
- **RESTRICT** for critical data (e.g., prevent plan deletion if subscriptions exist)

### Indexes
- All `created_at` and `updated_at` fields have indexes for time-based queries
- Foreign keys automatically have indexes
- Full-text search indexes on `title` and `description` fields

### Data Types
- **SERIAL** = Auto-incrementing integer (1, 2, 3...)
- **VARCHAR(n)** = Variable-length string, max n characters
- **TEXT** = Unlimited length text
- **INTEGER** = Whole number (-2B to 2B)
- **BIGINT** = Large integer (-9 quintillion to 9 quintillion)
- **DECIMAL(p,s)** = Fixed-point decimal (p digits, s after decimal)
- **BOOLEAN** = True/False
- **TIMESTAMP** = Date and time with timezone
- **DATE** = Date only
- **INET** = IP address (v4 or v6)
- **JSONB** = JSON data (binary, searchable)

### Naming Conventions
- Tables: Plural, lowercase, underscores (e.g., `video_views`)
- Columns: Singular, lowercase, underscores (e.g., `user_id`)
- Foreign keys: `{table}_id` (e.g., `user_id`)
- Indexes: `idx_{table}_{column(s)}` (e.g., `idx_users_email`)

### Query Performance
- Use indexes for WHERE, JOIN, ORDER BY columns
- JSONB columns support GIN indexes for fast JSON queries
- Partitioning recommended for `video_views`, `user_activity_log` (time-based)
- Consider materialized views for complex analytics queries

---

## 🔒 SECURITY CONSIDERATIONS

1. **Passwords:** Bcrypt hashed with salt, never stored plain text
2. **Tokens:** JWT tokens, session tokens hashed
3. **API Keys:** SHA256 hashed, never stored plain text
4. **Email:** Personal data, encrypted at rest recommended
5. **IP Addresses:** GDPR compliance required for EU users
6. **Payment Data:** Never store credit card numbers (use Stripe)

---

## 📈 MAINTENANCE

### Regular Tasks
- **Daily:** Update view counts, engagement scores
- **Weekly:** Calculate subscriber metrics, cleanup old sessions
- **Monthly:** Generate payout records, archive old logs
- **Quarterly:** Review indexes, optimize slow queries

### Backup Strategy
- **Full backup:** Daily at 2 AM UTC
- **Incremental:** Every 6 hours
- **Point-in-time recovery:** 30-day retention
- **Off-site backup:** AWS S3 or equivalent

---

**Total Tables:** 74  
**Total Indexes:** 150+  
**Total Foreign Keys:** 100+  
**Database Size (Estimated):** 50GB - 500GB depending on video count  

*Last Updated: October 22, 2025*  
*PostgreSQL Version: 14.x or higher*  
*Character Set: UTF-8*
