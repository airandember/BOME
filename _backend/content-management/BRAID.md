# 🧬 Content Management Braid - Backend
**Articles, tags, categories, and content organization**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **backend portion** of the Content Management Braid.  
> **Frontend portion**: See `_frontend/braids/content-management/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Backend Overview**

**Purpose**: Server-side content management, tagging, categorization, and publishing  
**Technology**: Go, PostgreSQL  
**Complexity**: Medium (CRUD, Hierarchies, Search)  
**Dependencies**: Auth Braid (authoring), User Mgmt Braid (permissions)

---

## 🎯 **Key Features**

### **1. Article Management**:
- Create, read, update, delete articles
- Draft and published states
- Author attribution
- Publish/unpublish workflow
- Article metadata (title, slug, content)

### **2. Tag System**:
- Tag creation and management
- Tag categories (e.g., "technology", "business")
- Tag assignment to content
- Tag suggestions
- Tag-based filtering

### **3. Category System**:
- Hierarchical categories
- Parent-child relationships
- Category-based organization
- Category browsing

### **4. Content Search**:
- Full-text search
- Tag-based search
- Category filtering
- Author filtering
- Date range filtering

### **5. SEO & Metadata**:
- Meta titles and descriptions
- Slugs for URLs
- Open Graph tags
- Schema.org markup
- Sitemap generation

---

## 🗄️ **Database Schema**

### **Articles Table** (Implemented):
```sql
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    excerpt TEXT,
    author_id INTEGER REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'draft', -- 'draft', 'published', 'archived'
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    views INTEGER DEFAULT 0,
    -- SEO fields
    meta_title VARCHAR(255),
    meta_description TEXT,
    og_image VARCHAR(500)
);

CREATE INDEX idx_articles_slug ON articles(slug);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_author_id ON articles(author_id);
CREATE INDEX idx_articles_published_at ON articles(published_at);
```

---

### **Tags Table** (Implemented):
**File**: `backend/migrations/*tags*.sql`

```sql
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(50), -- 'technology', 'business', 'lifestyle', etc.
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tags_slug ON tags(slug);
CREATE INDEX idx_tags_category ON tags(category);
CREATE INDEX idx_tags_usage_count ON tags(usage_count);
```

---

### **Tag Categories Table** (Implemented):
```sql
CREATE TABLE tag_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(50), -- Icon identifier
    color VARCHAR(20), -- Hex color code
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

### **Article Tags Junction Table** (Many-to-Many):
```sql
CREATE TABLE article_tags (
    article_id INTEGER REFERENCES articles(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (article_id, tag_id)
);

CREATE INDEX idx_article_tags_article_id ON article_tags(article_id);
CREATE INDEX idx_article_tags_tag_id ON article_tags(tag_id);
```

---

### **Categories Table** (Hierarchical):
```sql
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parent_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    level INTEGER DEFAULT 0, -- Depth in hierarchy
    path VARCHAR(500), -- Full path for easier queries
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_level ON categories(level);
```

---

### **Article Categories Junction Table**:
```sql
CREATE TABLE article_categories (
    article_id INTEGER REFERENCES articles(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT false, -- Primary category
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (article_id, category_id)
);

CREATE INDEX idx_article_categories_article_id ON article_categories(article_id);
CREATE INDEX idx_article_categories_category_id ON article_categories(category_id);
```

---

## 🌐 **API Endpoints**

### **Articles**:
```
GET    /api/v1/articles                # List articles (paginated)
GET    /api/v1/articles/:slug          # Get article by slug
POST   /api/v1/articles                # Create article (auth required)
PUT    /api/v1/articles/:id            # Update article (auth required)
DELETE /api/v1/articles/:id            # Delete article (auth required)
POST   /api/v1/articles/:id/publish    # Publish article
POST   /api/v1/articles/:id/unpublish  # Unpublish article
```

### **Tags**:
**File**: `backend/internal/routes/tags.go`

```
GET    /api/v1/tags                    # List all tags
GET    /api/v1/tags/:slug              # Get tag by slug
POST   /api/v1/tags                    # Create tag (admin)
PUT    /api/v1/tags/:id                # Update tag (admin)
DELETE /api/v1/tags/:id                # Delete tag (admin)
GET    /api/v1/tags/categories         # List tag categories
GET    /api/v1/tags/popular            # Get popular tags
```

### **Categories**:
```
GET    /api/v1/categories              # List categories (tree)
GET    /api/v1/categories/:slug        # Get category by slug
POST   /api/v1/categories              # Create category (admin)
PUT    /api/v1/categories/:id          # Update category (admin)
DELETE /api/v1/categories/:id          # Delete category (admin)
GET    /api/v1/categories/:id/children # Get child categories
```

### **Content Search**:
```
GET    /api/v1/search/articles?q=query&tags=tag1,tag2&category=slug
```

**Query Parameters**:
- `q` - Search query (full-text)
- `tags` - Comma-separated tag slugs
- `category` - Category slug
- `author` - Author ID
- `status` - Article status (published/draft)
- `from` - Start date
- `to` - End date
- `page` - Page number
- `limit` - Items per page (default 20)
- `sort` - Sort field (published_at, views, title)
- `order` - Sort order (asc, desc)

---

## 🔧 **Backend Services**

### **Tag Service** (`backend/internal/database/tags.go`):

**Key Operations**:
```go
// Tag CRUD
func GetAllTags() ([]Tag, error)
func GetTagBySlug(slug string) (*Tag, error)
func CreateTag(tag *Tag) error
func UpdateTag(tag *Tag) error
func DeleteTag(id int) error

// Tag usage
func IncrementTagUsage(tagID int) error
func GetPopularTags(limit int) ([]Tag, error)

// Tag categories
func GetTagCategories() ([]TagCategory, error)
func GetTagsByCategory(category string) ([]Tag, error)

// Article tagging
func AddTagToArticle(articleID, tagID int) error
func RemoveTagFromArticle(articleID, tagID int) error
func GetArticleTags(articleID int) ([]Tag, error)
func GetArticlesByTag(tagSlug string) ([]Article, error)
```

---

### **Article Service** (`backend/internal/routes/articles.go`):

**Key Operations**:
```go
// Article CRUD
func ListArticles(filters ArticleFilters) ([]Article, int, error)
func GetArticleBySlug(slug string) (*Article, error)
func CreateArticle(article *Article) error
func UpdateArticle(article *Article) error
func DeleteArticle(id int) error

// Publishing workflow
func PublishArticle(id int) error
func UnpublishArticle(id int) error
func GetDraftArticles(authorID int) ([]Article, error)

// View tracking
func IncrementArticleViews(id int) error

// Search
func SearchArticles(query SearchQuery) ([]Article, int, error)
```

---

## 📊 **Content Organization**

### **Tag System Design**:

**Tag Categories** (Implemented in DB):
- `technology` - Tech-related tags
- `business` - Business and entrepreneurship
- `design` - Design and UX
- `development` - Software development
- `lifestyle` - Lifestyle and personal
- `marketing` - Marketing and sales

**Popular Tags**:
```go
func GetPopularTags(limit int) ([]Tag, error) {
    query := `
        SELECT t.*, COUNT(at.article_id) as article_count
        FROM tags t
        LEFT JOIN article_tags at ON t.id = at.tag_id
        GROUP BY t.id
        ORDER BY article_count DESC, t.usage_count DESC
        LIMIT $1
    `
    // Returns most used tags
}
```

---

### **Category Hierarchy**:

**Example Structure**:
```
Technology
├── Web Development
│   ├── Frontend
│   └── Backend
├── Mobile Development
│   ├── iOS
│   └── Android
└── Data Science
    ├── Machine Learning
    └── Data Analytics

Business
├── Entrepreneurship
├── Marketing
└── Finance
```

**Hierarchical Queries**:
```go
func GetCategoryTree(parentID *int) ([]Category, error) {
    // Recursive query to get full tree
}

func GetCategoryPath(categoryID int) ([]Category, error) {
    // Get breadcrumb path to root
}
```

---

## 🔍 **Search Implementation**

### **Full-Text Search** (PostgreSQL):
```sql
-- Add full-text search index
ALTER TABLE articles ADD COLUMN search_vector tsvector;

CREATE INDEX idx_articles_search 
ON articles USING gin(search_vector);

-- Update trigger
CREATE FUNCTION articles_search_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.excerpt,'')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.content,'')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER articles_search_vector_update
BEFORE INSERT OR UPDATE ON articles
FOR EACH ROW EXECUTE FUNCTION articles_search_update();
```

**Search Query**:
```go
func SearchArticles(query string) ([]Article, error) {
    sql := `
        SELECT *
        FROM articles
        WHERE search_vector @@ to_tsquery('english', $1)
        AND status = 'published'
        ORDER BY ts_rank(search_vector, to_tsquery('english', $1)) DESC
    `
    // Returns ranked results
}
```

---

## 🔒 **Access Control**

### **Content Permissions**:

**Public Access**:
- ✅ View published articles
- ✅ Browse categories and tags
- ✅ Search published content

**Authenticated Users**:
- ✅ Create draft articles
- ✅ Edit own articles
- ✅ Delete own articles

**Admins/Editors**:
- ✅ Publish/unpublish any article
- ✅ Edit any article
- ✅ Delete any article
- ✅ Manage tags and categories
- ✅ Moderate content

**Middleware**:
```go
// Check if user can edit article
func CanEditArticle(userID, articleID int) bool {
    article := GetArticle(articleID)
    user := GetUser(userID)
    
    return user.IsAdmin() || article.AuthorID == userID
}
```

---

## ⚡ **Performance Optimizations**

### **Caching Strategy**:
- **Published articles**: 5-minute cache
- **Tag lists**: 10-minute cache
- **Category tree**: 15-minute cache
- **Popular tags**: 30-minute cache
- **Search results**: 2-minute cache

### **Database Indexes**:
- ✅ Article slugs (unique lookup)
- ✅ Tag slugs (fast tag queries)
- ✅ Category hierarchy (tree traversal)
- ✅ Full-text search (search performance)
- ✅ Junction tables (relationship queries)

### **Pagination**:
- Article lists: 20 per page
- Tag lists: 50 per page
- Search results: 10 per page

---

## 📈 **Content Analytics** (Future)

**Metrics to Track**:
- Article views
- Read time
- Popular tags
- Category popularity
- Author performance
- Search queries

---

## 🧬 **Strands (Complete Flows)**

### **1. Content Publishing Strand**:
Complete flow from draft creation to publication

### **2. Tag System Strand**:
Tag creation, assignment, and filtering

### **3. Category Management Strand**:
Hierarchical category organization

---

## 📝 **Known Technical Debt**

### **Current Limitations**:
1. No rich text editor (plain text/markdown)
2. No media upload for articles
3. No article revisions/version history
4. No scheduled publishing
5. No content workflow (draft → review → publish)
6. Limited SEO metadata

### **Future Enhancements**:
1. ✅ Rich text editor (WYSIWYG)
2. ✅ Image upload and management
3. ✅ Article version history
4. ✅ Scheduled publishing
5. ✅ Editorial workflow
6. ✅ Advanced SEO tools
7. ✅ Content recommendations
8. ✅ Related articles algorithm

---

## 🔗 **Related Braids**

### **Depends On**:
- **Authentication Braid**: Author identity
- **User Management Braid**: Author profiles, permissions

### **Consumed By**:
- **Admin Dashboard Braid**: Content management UI
- **Analytics Braid**: Content performance
- **Video Streaming Braid**: Video metadata

---

## 🚀 **Quick Start**

### **Understanding Content Management** (10 min):
1. Read this BRAID.md (5 min)
2. Check `backend/internal/routes/tags.go` (3 min)
3. Review tag system strand (2 min)

### **Creating Content**:
```go
// Create article
article := &Article{
    Title:   "My Article",
    Slug:    "my-article",
    Content: "Article content...",
    AuthorID: userID,
    Status:  "draft",
}
database.CreateArticle(article)

// Add tags
database.AddTagToArticle(article.ID, tagID)

// Publish
database.PublishArticle(article.ID)
```

---

**Last Updated**: October 14, 2025  
**Status**: Core implementation, enhancements planned  
**Technology**: Go + PostgreSQL  
**Frontend Counterpart**: `_frontend/braids/content-management/`

---

**Navigate**:  
[🏠 Master Index](../../BRAIDS_INDEX.md) | [🎨 Frontend Braid](../../_frontend/braids/content-management/BRAID.md)

