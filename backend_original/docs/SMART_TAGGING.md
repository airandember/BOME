# 🏷️ Smart Video Tagging System

## Overview

The Smart Video Tagging System automatically generates meaningful, searchable tags from video titles using intelligent text analysis. It implements a sophisticated algorithm that extracts names, removes common articles, and categorizes content by subject area.

## 🎯 How It Works

### 1. Title Analysis Algorithm

The system follows these steps:

1. **Extract Name**: Takes characters before the first "-" as the presenter's name
2. **Process Remaining Text**: Removes common articles (and, the, of, for, a, an, etc.)
3. **Generate Tags**: Creates individual tags from meaningful remaining words
4. **Update Database**: Sends tag array to database and sets `tagged = true`
5. **Track Analytics**: Maintains word frequency and categorization data

### 2. Example Processing

**Input Title**: `"John Smith - The Ancient Ruins of Mesoamerica and Archaeological Discoveries"`

**Processing Steps**:
- Name: `"John Smith"`
- Remaining: `"The Ancient Ruins of Mesoamerica and Archaeological Discoveries"`
- Filtered: `"Ancient Ruins Mesoamerica Archaeological Discoveries"`

**Final Tags**: `["John Smith", "Ancient", "Ruins", "Mesoamerica", "Archaeological", "Discoveries"]`

## 🗄️ Database Schema

### New Tables

#### `video_tags`
- `id`: Primary key
- `word`: Tag word (unique)
- `frequency`: Usage count
- `category_id`: Reference to tag_categories
- `created_at`, `updated_at`: Timestamps

#### `tag_categories`
- `id`: Primary key
- `name`: Category name (unique)
- `description`: Category description
- `color`: Hex color code for UI
- `created_at`, `updated_at`: Timestamps

### Updated Tables

#### `master_video_list`
- Added `tagged` boolean column (default: false)
- Added index on `tagged` column for performance

## 🚀 API Endpoints

### Smart Tagging

#### `POST /api/v1/admin/master-videos/:id/auto-tag`
Auto-generates tags for a specific video.

**Response**:
```json
{
  "success": true,
  "message": "Video tagged successfully",
  "result": {
    "name": "John Smith",
    "tags": ["John Smith", "Ancient", "Ruins", "Mesoamerica"],
    "original_title": "John Smith - The Ancient Ruins of Mesoamerica",
    "processed_title": "Ancient Ruins Mesoamerica"
  }
}
```

### Tag Analytics

#### `GET /api/v1/admin/master-videos/tags/analytics`
Returns comprehensive tag analytics and statistics.

**Response**:
```json
{
  "success": true,
  "data": {
    "tag_frequency": [
      {"word": "archaeology", "frequency": 45},
      {"word": "mesoamerica", "frequency": 32}
    ],
    "categories": [
      {"name": "Archaeology", "description": "Archaeological terms", "color": "#8b5cf6"}
    ],
    "total_videos": 150,
    "tagged_videos": 120,
    "untagged_videos": 30,
    "tagging_percentage": 80.0
  }
}
```

#### `GET /api/v1/admin/master-videos/tags/untagged`
Returns videos that haven't been tagged yet.

**Query Parameters**:
- `limit`: Maximum number of videos to return (default: 50)

**Response**:
```json
{
  "success": true,
  "videos": [...],
  "count": 30
}
```

## 🎨 Tag Categories

The system automatically categorizes tags into these subject areas:

| Category | Description | Color |
|----------|-------------|-------|
| Archaeology | Archaeological terms and concepts | #8b5cf6 |
| Geography | Geographic locations and features | #06b6d4 |
| DNA Research | Genetic and DNA-related terms | #10b981 |
| Linguistics | Language and linguistic terms | #f59e0b |
| Historical Evidence | Historical documentation | #ef4444 |
| Cultural Studies | Cultural and anthropological terms | #ec4899 |
| Religious Studies | Religious and theological terms | #6366f1 |
| Documentary | Documentary and media terms | #84cc16 |
| Lecture | Educational and lecture terms | #f97316 |
| Interview | Interview and discussion terms | #06b6d4 |
| Presentation | Presentation and visual terms | #8b5cf6 |
| Virtual Tour | Tour and exploration terms | #10b981 |

## 🔧 Implementation Details

### Smart Tagging Service

The core logic is implemented in `services/smart_tagging.go`:

- **Article Filtering**: Comprehensive list of 50+ common articles and prepositions
- **Word Cleaning**: Removes punctuation and normalizes case
- **Duplicate Prevention**: Ensures no duplicate tags within a video
- **Fallback Logic**: Handles titles without dashes gracefully

### Database Integration

- **Tag Updates**: Atomic updates with JSON serialization
- **Frequency Tracking**: Incremental updates with conflict resolution
- **Performance**: Indexed queries for fast retrieval
- **Audit Trail**: Timestamps for all operations

## 📊 Usage Examples

### Frontend Integration

```typescript
// Auto-tag a video
const response = await fetch(`/api/v1/admin/master-videos/${videoId}/auto-tag`, {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
});

const result = await response.json();
console.log('Generated tags:', result.result.tags);
```

### Batch Processing

```typescript
// Get untagged videos
const response = await fetch('/api/v1/admin/master-videos/tags/untagged?limit=100');
const { videos } = await response.json();

// Tag each video
for (const video of videos) {
  await fetch(`/api/v1/admin/master-videos/${video.id}/auto-tag`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
}
```

## 🧪 Testing

Run the test script to see the tagging system in action:

```bash
cd backend
go run cmd/test-tagging/main.go
```

This will demonstrate:
- Title processing with various formats
- Tag generation and categorization
- Available tag categories

## 🔮 Future Enhancements

### Planned Features

1. **Machine Learning**: Improve categorization accuracy
2. **Custom Categories**: User-defined tag categories
3. **Tag Relationships**: Semantic connections between tags
4. **Bulk Operations**: Tag multiple videos simultaneously
5. **Tag Suggestions**: AI-powered tag recommendations

### Analytics Dashboard

- Tag frequency visualization
- Category distribution charts
- Trending tags over time
- Search and filter capabilities
- Export functionality

## 📝 Migration

To add the smart tagging system to an existing database:

1. **Run Migration**: Execute `migrations/007_add_tagged_column.sql`
2. **Update Backend**: Deploy the new code with smart tagging endpoints
3. **Test Integration**: Verify API endpoints are working
4. **Monitor Performance**: Check database performance with new indexes

## 🎯 Benefits

- **Improved Search**: Better content discovery through meaningful tags
- **Content Organization**: Automatic categorization by subject area
- **Analytics Insights**: Track popular topics and trends
- **User Experience**: Enhanced filtering and recommendation systems
- **Content Management**: Streamlined video organization workflow

---

*The Smart Video Tagging System transforms unstructured video titles into organized, searchable metadata, making your video library more discoverable and manageable.*
