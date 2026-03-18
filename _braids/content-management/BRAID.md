# Braid: content-management

**Architecture:** Full-Stack Braid (Frontend to Backend)
**Last Updated:** 2025-10-17

---

## Backend Architecture

**Articles, tags, categories, and content organization**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **backend portion** of the Content Management Braid.  
> **Frontend portion**: See `_frontend/braids/content-management/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Backend Overview**

**Purpose**: Server-side content management, tagging, categorization, and publishing  
**Technology**: Go, PostgreSQL  
**Complexity**: Medium (CRUD, Hierarchies, Search)  
**Dependencies**: Auth Braid (authoring), User Mgmt Braid (permissions)

---

## **File Map** (Production Code)

| Layer | Production Path | Description |
|-------|-----------------|-------------|
| Handlers | `backend/content/handlers/tags.go` | Tag CRUD API handlers |
| Models | `backend/content/models/tags.go` | Tag/category data models |
| Routes | `backend/internal/routes/tags.go`, `backend/internal/routes/articles.go` | Tag and article API routes |
| Database | `backend/internal/database/` | Migrations for articles, tags, categories |

**Frontend:** `frontend/src/routes/articles/`, `frontend/src/lib/` (tag components)

---

## **Key Features**

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

## ðŸ—„ï¸ **Database Schema**

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

## ðŸŒ **API Endpoints**

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

## ðŸ”§ **Backend Services**

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

## ðŸ“Š **Content Organization**

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
â”œâ”€â”€ Web Development
â”‚   â”œâ”€â”€ Frontend
â”‚   â””â”€â”€ Backend
â”œâ”€â”€ Mobile Development
â”‚   â”œâ”€â”€ iOS
â”‚   â””â”€â”€ Android
â””â”€â”€ Data Science
    â”œâ”€â”€ Machine Learning
    â””â”€â”€ Data Analytics

Business
â”œâ”€â”€ Entrepreneurship
â”œâ”€â”€ Marketing
â””â”€â”€ Finance
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

## ðŸ” **Search Implementation**

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

## ðŸ”’ **Access Control**

### **Content Permissions**:

**Public Access**:
- âœ… View published articles
- âœ… Browse categories and tags
- âœ… Search published content

**Authenticated Users**:
- âœ… Create draft articles
- âœ… Edit own articles
- âœ… Delete own articles

**Admins/Editors**:
- âœ… Publish/unpublish any article
- âœ… Edit any article
- âœ… Delete any article
- âœ… Manage tags and categories
- âœ… Moderate content

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

## âš¡ **Performance Optimizations**

### **Caching Strategy**:
- **Published articles**: 5-minute cache
- **Tag lists**: 10-minute cache
- **Category tree**: 15-minute cache
- **Popular tags**: 30-minute cache
- **Search results**: 2-minute cache

### **Database Indexes**:
- âœ… Article slugs (unique lookup)
- âœ… Tag slugs (fast tag queries)
- âœ… Category hierarchy (tree traversal)
- âœ… Full-text search (search performance)
- âœ… Junction tables (relationship queries)

### **Pagination**:
- Article lists: 20 per page
- Tag lists: 50 per page
- Search results: 10 per page

---

## ðŸ“ˆ **Content Analytics** (Future)

**Metrics to Track**:
- Article views
- Read time
- Popular tags
- Category popularity
- Author performance
- Search queries

---

## ðŸ§¬ **Strands (Complete Flows)**

### **1. Content Publishing Strand**:
Complete flow from draft creation to publication

### **2. Tag System Strand**:
Tag creation, assignment, and filtering

### **3. Category Management Strand**:
Hierarchical category organization

---

## ðŸ“ **Known Technical Debt**

### **Current Limitations**:
1. No rich text editor (plain text/markdown)
2. No media upload for articles
3. No article revisions/version history
4. No scheduled publishing
5. No content workflow (draft â†’ review â†’ publish)
6. Limited SEO metadata

### **Future Enhancements**:
1. âœ… Rich text editor (WYSIWYG)
2. âœ… Image upload and management
3. âœ… Article version history
4. âœ… Scheduled publishing
5. âœ… Editorial workflow
6. âœ… Advanced SEO tools
7. âœ… Content recommendations
8. âœ… Related articles algorithm

---

## ðŸ”— **Related Braids**

### **Depends On**:
- **Authentication Braid**: Author identity
- **User Management Braid**: Author profiles, permissions

### **Consumed By**:
- **Admin Dashboard Braid**: Content management UI
- **Analytics Braid**: Content performance
- **Video Streaming Braid**: Video metadata

---

## ðŸš€ **Quick Start**

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
[ðŸ  Master Index](../../BRAIDS_INDEX.md) | [ðŸŽ¨ Frontend Braid](../../_frontend/braids/content-management/BRAID.md)



---

## Frontend Architecture

**Svelte5 UI for articles, tags, and content organization**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **frontend portion** of the Content Management Braid.  
> **Backend portion**: See `_braids/content-management/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Frontend Overview**

**Purpose**: User interface for content browsing, creation, and management  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/articles`, `/articles/[slug]`, `/admin/tags-categories`  
**State Management**: Svelte stores for content state

---

## ðŸŽ¯ **Key Features**

### **1. Article Browsing**:
- Article listing with pagination
- Tag-based filtering
- Category-based filtering
- Search functionality
- Sort options (newest, popular, title)

### **2. Article Reading**:
- Clean reading experience
- Author information
- Related articles
- Tag display
- Social sharing
- Reading time estimate

### **3. Content Creation** (Admin):
- Article editor
- Markdown support
- Tag selection
- Category selection
- SEO metadata editor
- Draft/publish workflow

### **4. Tag Management** (Admin):
- Browse all tags
- Create/edit/delete tags
- Tag categories
- Tag usage statistics
- Bulk operations

### **5. Category Management** (Admin):
- Hierarchical category tree
- Create/edit/delete categories
- Drag-and-drop reordering
- Category assignment

---

## ðŸ“„ **Frontend Pages**

### **1. Articles List Page** (`/articles`)
**File**: `frontend/src/routes/articles/+page.svelte`

**Features**:
- Grid of article cards
- Tag filter sidebar
- Category breadcrumbs
- Search bar
- Pagination
- Sort dropdown
- Loading states

**Example UI**:
```svelte
<div class="articles-page">
  <div class="sidebar">
    <h3>Filter by Tags</h3>
    <TagFilter bind:selectedTags />
    
    <h3>Categories</h3>
    <CategoryTree bind:selectedCategory />
  </div>
  
  <div class="main-content">
    <SearchBar bind:query />
    
    <div class="articles-grid">
      {#each articles as article}
        <ArticleCard {article} />
      {/each}
    </div>
    
    <Pagination bind:page {totalPages} />
  </div>
</div>
```

---

### **2. Article Detail Page** (`/articles/[slug]`)
**File**: `frontend/src/routes/articles/[slug]/+page.svelte`

**Features**:
- Article content (rendered markdown)
- Author bio card
- Published date
- Reading time
- Tag list (clickable)
- Share buttons
- Related articles
- Comments (future)

**Example UI**:
```svelte
<article class="article-detail">
  <header>
    <h1>{article.title}</h1>
    <div class="meta">
      <AuthorCard author={article.author} />
      <time>{formatDate(article.published_at)}</time>
      <span>{article.read_time} min read</span>
    </div>
  </header>
  
  <div class="content">
    {@html renderMarkdown(article.content)}
  </div>
  
  <footer>
    <TagList tags={article.tags} />
    <ShareButtons url={articleUrl} />
  </footer>
  
  <aside>
    <h3>Related Articles</h3>
    <RelatedArticles {article} />
  </aside>
</article>
```

---

### **3. Admin Tags & Categories** (`/admin/tags-categories`)
**File**: `frontend/src/routes/admin/tags-categories/+page.svelte`

**Features**:
- Two-panel layout (tags | categories)
- Create new tag form
- Create new category form
- Edit/delete buttons
- Tag categories dropdown
- Category hierarchy tree
- Usage statistics
- Bulk selection and actions

**Example UI**:
```svelte
<div class="admin-content-management">
  <div class="tags-panel">
    <h2>Tags</h2>
    
    <button on:click={() => showCreateTag = true}>
      + New Tag
    </button>
    
    <TagCategoryFilter bind:selectedCategory />
    
    <table class="tags-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Category</th>
          <th>Articles</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each filteredTags as tag}
          <tr>
            <td>{tag.name}</td>
            <td>{tag.category}</td>
            <td>{tag.usage_count}</td>
            <td>
              <button on:click={() => editTag(tag)}>Edit</button>
              <button on:click={() => deleteTag(tag)}>Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  
  <div class="categories-panel">
    <h2>Categories</h2>
    
    <button on:click={() => showCreateCategory = true}>
      + New Category
    </button>
    
    <CategoryTree
      categories={categories}
      onEdit={editCategory}
      onDelete={deleteCategory}
      onReorder={reorderCategories}
    />
  </div>
</div>
```

---

## ðŸ§© **Frontend Components**

### **ArticleCard Component**
**Purpose**: Display article preview

**Props**:
```typescript
interface ArticleCardProps {
  article: Article;
  showAuthor?: boolean;
  showTags?: boolean;
}
```

**Features**:
- Article thumbnail
- Title and excerpt
- Author info
- Tag chips
- Read time
- Click to view

---

### **TagFilter Component**
**Purpose**: Filter content by tags

**Features**:
- Tag list with checkboxes
- Tag categories
- Search tags
- Selected tags display
- Clear all button

---

### **CategoryTree Component**
**Purpose**: Hierarchical category navigation

**Features**:
- Expandable/collapsible tree
- Click to filter
- Active category highlight
- Category icons
- Drag-and-drop (admin)

---

### **ArticleEditor Component** (Admin)
**Purpose**: Create/edit articles

**Features**:
- Markdown editor
- Live preview
- Title and slug fields
- Tag selector (multi-select)
- Category selector
- SEO metadata fields
- Save draft / Publish buttons
- Auto-save

**Example**:
```svelte
<div class="article-editor">
  <input bind:value={article.title} placeholder="Article Title" />
  <input bind:value={article.slug} placeholder="url-slug" />
  
  <div class="editor-layout">
    <div class="markdown-editor">
      <textarea bind:value={article.content} />
    </div>
    <div class="preview">
      {@html renderMarkdown(article.content)}
    </div>
  </div>
  
  <div class="metadata">
    <TagSelector bind:selectedTags />
    <CategorySelector bind:selectedCategory />
    
    <input bind:value={article.meta_title} placeholder="SEO Title" />
    <textarea bind:value={article.meta_description} placeholder="SEO Description" />
  </div>
  
  <div class="actions">
    <button on:click={saveDraft}>Save Draft</button>
    <button on:click={publish} class="primary">Publish</button>
  </div>
</div>
```

---

## ðŸ—ƒï¸ **Frontend Stores**

### **Articles Store** (`$lib/stores/articles.ts`)
**Purpose**: Manage article state

**State**:
```typescript
interface ArticlesState {
  articles: Article[];
  currentArticle: Article | null;
  total: number;
  page: number;
  loading: boolean;
  error: string | null;
  filters: ArticleFilters;
}

interface ArticleFilters {
  query?: string;
  tags?: string[];
  category?: string;
  author?: string;
  status?: 'published' | 'draft';
  sort?: 'newest' | 'popular' | 'title';
}
```

**Methods**:
```typescript
export const articles = {
  async loadArticles(filters: ArticleFilters) {
    // GET /api/v1/articles
  },
  
  async getArticle(slug: string) {
    // GET /api/v1/articles/:slug
  },
  
  async createArticle(article: Article) {
    // POST /api/v1/articles
  },
  
  async updateArticle(id: number, article: Article) {
    // PUT /api/v1/articles/:id
  },
  
  async deleteArticle(id: number) {
    // DELETE /api/v1/articles/:id
  },
  
  async publishArticle(id: number) {
    // POST /api/v1/articles/:id/publish
  },
  
  setFilters(filters: ArticleFilters) {
    // Update and reload
  }
};
```

---

### **Tags Store** (`$lib/stores/tags.ts`)
**Purpose**: Manage tags state

**State**:
```typescript
interface TagsState {
  tags: Tag[];
  tagCategories: TagCategory[];
  popularTags: Tag[];
  loading: boolean;
}
```

**Methods**:
```typescript
export const tags = {
  async loadTags() {
    // GET /api/v1/tags
  },
  
  async createTag(tag: Tag) {
    // POST /api/v1/tags
  },
  
  async updateTag(id: number, tag: Tag) {
    // PUT /api/v1/tags/:id
  },
  
  async deleteTag(id: number) {
    // DELETE /api/v1/tags/:id
  },
  
  async loadPopularTags() {
    // GET /api/v1/tags/popular
  }
};
```

---

### **Categories Store** (`$lib/stores/categories.ts`)
**Purpose**: Manage category state

**State**:
```typescript
interface CategoriesState {
  categories: Category[];
  categoryTree: CategoryNode[];
  loading: boolean;
}
```

**Methods**:
```typescript
export const categories = {
  async loadCategories() {
    // GET /api/v1/categories
  },
  
  async createCategory(category: Category) {
    // POST /api/v1/categories
  },
  
  async updateCategory(id: number, category: Category) {
    // PUT /api/v1/categories/:id
  },
  
  async deleteCategory(id: number) {
    // DELETE /api/v1/categories/:id
  },
  
  buildTree(categories: Category[]): CategoryNode[] {
    // Build hierarchical tree structure
  }
};
```

---

## ðŸŽ¨ **UI Patterns**

### **Tag Filtering**:
```svelte
<script>
  import { articles, tags } from '$lib/stores';
  
  let selectedTags = [];
  
  $: filteredArticles = $articles.articles;
  
  async function filterByTags() {
    await articles.setFilters({
      tags: selectedTags.map(t => t.slug)
    });
  }
</script>

<div class="tag-filter">
  {#each $tags.tags as tag}
    <label>
      <input
        type="checkbox"
        value={tag.slug}
        bind:group={selectedTags}
        on:change={filterByTags}
      />
      {tag.name} ({tag.usage_count})
    </label>
  {/each}
</div>
```

---

### **Markdown Rendering**:
```typescript
import { marked } from 'marked';

export function renderMarkdown(content: string): string {
  return marked(content, {
    breaks: true,
    gfm: true,
    sanitize: true // Prevent XSS
  });
}
```

---

### **Reading Time Calculation**:
```typescript
export function calculateReadTime(content: string): number {
  const wordsPerMinute = 200;
  const words = content.trim().split(/\s+/).length;
  return Math.ceil(words / wordsPerMinute);
}
```

---

## ðŸ”„ **Data Flow Examples**

### **Browse Articles**:
```
1. User visits /articles
2. Component loads articles from store
3. Store.loadArticles() called
4. API: GET /api/v1/articles?page=1
5. Backend returns articles + total
6. Store updates state
7. UI renders article cards
8. Pagination displayed
```

### **Filter by Tag**:
```
1. User clicks tag
2. Selected tags updated
3. Store.setFilters({ tags: ['tag-slug'] })
4. API: GET /api/v1/articles?tags=tag-slug
5. Backend filters and returns
6. Store updates articles
7. UI re-renders with filtered results
```

### **Create Article** (Admin):
```
1. Admin opens editor
2. Fills in title, content, tags
3. Clicks "Publish"
4. Store.createArticle(article)
5. API: POST /api/v1/articles
6. Backend creates article
7. Store updates with new article
8. Redirect to article page
9. Success notification
```

---

## ðŸ”’ **Security**

### **Content Sanitization**:
- âœ… Markdown sanitized before rendering
- âœ… XSS prevention in user content
- âœ… HTML entity encoding

### **Access Control**:
- âœ… Editor UI hidden for non-admins
- âœ… API calls require auth
- âœ… Draft articles only visible to authors

---

## ðŸ“Š **Performance**

### **Optimizations**:
- **Lazy loading**: Images lazy loaded
- **Code splitting**: Editor loaded on demand
- **Caching**: Article list cached
- **Virtual scrolling**: For long article lists
- **Debounced search**: 300ms delay

---

## ðŸŽ¯ **Accessibility**

### **Features**:
- âœ… Semantic HTML
- âœ… ARIA labels
- âœ… Keyboard navigation
- âœ… Screen reader support
- âœ… Focus management
- âœ… Color contrast compliance

---

## ðŸ“ **Known Issues**

### **To Implement**:
1. Rich text editor (WYSIWYG)
2. Image upload in editor
3. Article version history UI
4. Scheduled publishing UI
5. Advanced search filters
6. Related articles algorithm

---

## ðŸš€ **Quick Links**

**Actual Files**:
- Articles List: `frontend/src/routes/articles/+page.svelte`
- Article Detail: `frontend/src/routes/articles/[slug]/+page.svelte`
- Admin Tags: `frontend/src/routes/admin/tags-categories/+page.svelte`
- Articles Store: `frontend/src/lib/stores/articles.ts` (to create)
- Tags Store: `frontend/src/lib/stores/tags.ts` (to create)

---

**Last Updated**: October 14, 2025  
**Status**: Core structure defined, implementation in progress  
**Technology**: Svelte 5 + TypeScript  
**Backend Counterpart**: `_braids/content-management/backend/`

---

**Navigate**:  
[ðŸ  Master Index](../../../BRAIDS_INDEX.md) | [â¬…ï¸ Backend Braid](../../_braids/content-management/backend/BRAID.md)



---

## Integration Notes

- Frontend: `_braids/content-management/frontend/`
- Backend: `_braids/content-management/backend/`

This braid represents a complete vertical slice of functionality.

