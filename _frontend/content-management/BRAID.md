# 🧬 Content Management Braid - Frontend
**Svelte5 UI for articles, tags, and content organization**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **frontend portion** of the Content Management Braid.  
> **Backend portion**: See `_braids/content-management/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Frontend Overview**

**Purpose**: User interface for content browsing, creation, and management  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/articles`, `/articles/[slug]`, `/admin/tags-categories`  
**State Management**: Svelte stores for content state

---

## 🎯 **Key Features**

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

## 📄 **Frontend Pages**

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

## 🧩 **Frontend Components**

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

## 🗃️ **Frontend Stores**

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

## 🎨 **UI Patterns**

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

## 🔄 **Data Flow Examples**

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

## 🔒 **Security**

### **Content Sanitization**:
- ✅ Markdown sanitized before rendering
- ✅ XSS prevention in user content
- ✅ HTML entity encoding

### **Access Control**:
- ✅ Editor UI hidden for non-admins
- ✅ API calls require auth
- ✅ Draft articles only visible to authors

---

## 📊 **Performance**

### **Optimizations**:
- **Lazy loading**: Images lazy loaded
- **Code splitting**: Editor loaded on demand
- **Caching**: Article list cached
- **Virtual scrolling**: For long article lists
- **Debounced search**: 300ms delay

---

## 🎯 **Accessibility**

### **Features**:
- ✅ Semantic HTML
- ✅ ARIA labels
- ✅ Keyboard navigation
- ✅ Screen reader support
- ✅ Focus management
- ✅ Color contrast compliance

---

## 📝 **Known Issues**

### **To Implement**:
1. Rich text editor (WYSIWYG)
2. Image upload in editor
3. Article version history UI
4. Scheduled publishing UI
5. Advanced search filters
6. Related articles algorithm

---

## 🚀 **Quick Links**

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
[🏠 Master Index](../../../BRAIDS_INDEX.md) | [⬅️ Backend Braid](../../_braids/content-management/backend/BRAID.md)

