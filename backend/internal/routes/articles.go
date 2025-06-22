package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Article represents a blog article
type ArticleStruct struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Slug        string       `json:"slug"`
	Content     string       `json:"content"`
	Excerpt     string       `json:"excerpt"`
	FeaturedImg string       `json:"featuredImg"`
	CategoryID  int          `json:"categoryId"`
	Category    string       `json:"category"`
	AuthorID    int          `json:"authorId"`
	Author      AuthorStruct `json:"author"`
	Tags        []string     `json:"tags"`
	Featured    bool         `json:"featured"`
	Published   bool         `json:"published"`
	ViewCount   int          `json:"viewCount"`
	ReadTime    int          `json:"readTime"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`
}

// AuthorStruct represents an article author
type AuthorStruct struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	Verified bool   `json:"verified"`
}

// ArticleCategory represents an article category
type ArticleCategory struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	ArticleCount int    `json:"articleCount"`
}

// MOCK AUTHORS DATA
var MOCK_AUTHORS = []AuthorStruct{
	{
		ID:       1,
		Name:     "Dr. Michael Richardson",
		Email:    "m.richardson@byu.edu",
		Bio:      "Professor of Ancient Studies at Brigham Young University, specializing in Mesoamerican archaeology and Book of Mormon geography.",
		Avatar:   "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		Role:     "Professor of Ancient Studies",
		Verified: true,
	},
	{
		ID:       2,
		Name:     "Sarah Chen",
		Email:    "s.chen@byu.edu",
		Bio:      "Research Associate in Linguistics, focusing on ancient Hebrew and Egyptian language patterns in religious texts.",
		Avatar:   "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		Role:     "Research Associate in Linguistics",
		Verified: true,
	},
	{
		ID:       3,
		Name:     "Dr. James Peterson",
		Email:    "j.peterson@byu.edu",
		Bio:      "Associate Professor of History, specializing in 19th-century American religious movements and early Mormon history.",
		Avatar:   "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		Role:     "Associate Professor of History",
		Verified: true,
	},
}

// MOCK ARTICLE CATEGORIES
var MOCK_ARTICLE_CATEGORIES = []ArticleCategory{
	{ID: 1, Name: "Archaeology", Slug: "archaeology", Description: "Archaeological discoveries and evidence", Color: "#8B5A2B", ArticleCount: 4},
	{ID: 2, Name: "Scripture Study", Slug: "scripture-study", Description: "In-depth analysis of Book of Mormon passages", Color: "#1E40AF", ArticleCount: 3},
	{ID: 3, Name: "Historical Context", Slug: "historical-context", Description: "Historical background and context", Color: "#059669", ArticleCount: 3},
	{ID: 4, Name: "Linguistics", Slug: "linguistics", Description: "Language analysis and patterns", Color: "#7C2D12", ArticleCount: 2},
	{ID: 5, Name: "Geography", Slug: "geography", Description: "Geographic studies and theories", Color: "#B45309", ArticleCount: 2},
	{ID: 6, Name: "Comparative Religion", Slug: "comparative-religion", Description: "Comparisons with other religious traditions", Color: "#7C3AED", ArticleCount: 2},
	{ID: 7, Name: "Testimonies", Slug: "testimonies", Description: "Personal testimonies and spiritual insights", Color: "#DC2626", ArticleCount: 1},
	{ID: 8, Name: "Academic Research", Slug: "academic-research", Description: "Scholarly research and peer-reviewed studies", Color: "#374151", ArticleCount: 1},
}

// MOCK ARTICLES DATA
var MOCK_ARTICLES = []ArticleStruct{
	{
		ID:          1,
		Title:       "Recent Archaeological Discoveries in Mesoamerica",
		Slug:        "recent-archaeological-discoveries-mesoamerica",
		Content:     "Recent excavations in the Yucatan Peninsula have uncovered remarkable evidence of advanced civilizations that flourished during the timeframe described in the Book of Mormon...",
		Excerpt:     "Exploring groundbreaking archaeological findings that illuminate the world of the Book of Mormon.",
		FeaturedImg: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		CategoryID:  1,
		Category:    "Archaeology",
		AuthorID:    1,
		Author:      MOCK_AUTHORS[0],
		Tags:        []string{"archaeology", "mesoamerica", "evidence", "civilization"},
		Featured:    true,
		Published:   true,
		ViewCount:   4567,
		ReadTime:    12,
		CreatedAt:   "2024-01-15T10:30:00Z",
		UpdatedAt:   "2024-01-15T10:30:00Z",
	},
	{
		ID:          2,
		Title:       "Hebrew Patterns in Book of Mormon Names",
		Slug:        "hebrew-patterns-book-mormon-names",
		Content:     "A detailed analysis of naming conventions in the Book of Mormon reveals striking parallels to ancient Hebrew naming patterns...",
		Excerpt:     "Examining the linguistic evidence for Hebrew influence in Book of Mormon nomenclature.",
		FeaturedImg: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		CategoryID:  4,
		Category:    "Linguistics",
		AuthorID:    2,
		Author:      MOCK_AUTHORS[1],
		Tags:        []string{"hebrew", "linguistics", "names", "ancient-languages"},
		Featured:    true,
		Published:   true,
		ViewCount:   3421,
		ReadTime:    8,
		CreatedAt:   "2024-01-18T14:20:00Z",
		UpdatedAt:   "2024-01-18T14:20:00Z",
	},
	{
		ID:          3,
		Title:       "Joseph Smith and the Translation Process",
		Slug:        "joseph-smith-translation-process",
		Content:     "Historical accounts of the Book of Mormon translation process provide insights into both the practical and spiritual aspects of this remarkable work...",
		Excerpt:     "Understanding the historical context and process of the Book of Mormon translation.",
		FeaturedImg: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		CategoryID:  3,
		Category:    "Historical Context",
		AuthorID:    3,
		Author:      MOCK_AUTHORS[2],
		Tags:        []string{"joseph-smith", "translation", "history", "revelation"},
		Featured:    false,
		Published:   true,
		ViewCount:   2890,
		ReadTime:    15,
		CreatedAt:   "2024-01-20T09:45:00Z",
		UpdatedAt:   "2024-01-20T09:45:00Z",
	},
}

// ARTICLES ENDPOINTS

// GetArticlesHandler returns paginated articles with optional filtering
func GetArticlesHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	search := c.Query("search")
	featured := c.Query("featured")

	filteredArticles := MOCK_ARTICLES

	// Apply category filter
	if category != "" {
		filtered := []ArticleStruct{}
		for _, article := range filteredArticles {
			if strings.EqualFold(article.Category, category) {
				filtered = append(filtered, article)
			}
		}
		filteredArticles = filtered
	}

	// Apply search filter
	if search != "" {
		searchLower := strings.ToLower(search)
		filtered := []ArticleStruct{}
		for _, article := range filteredArticles {
			if strings.Contains(strings.ToLower(article.Title), searchLower) ||
				strings.Contains(strings.ToLower(article.Content), searchLower) ||
				strings.Contains(strings.ToLower(article.Excerpt), searchLower) {
				filtered = append(filtered, article)
			}
		}
		filteredArticles = filtered
	}

	// Apply featured filter
	if featured == "true" {
		filtered := []ArticleStruct{}
		for _, article := range filteredArticles {
			if article.Featured {
				filtered = append(filtered, article)
			}
		}
		filteredArticles = filtered
	}

	// Apply pagination
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if startIndex >= len(filteredArticles) {
		filteredArticles = []ArticleStruct{}
	} else if endIndex > len(filteredArticles) {
		filteredArticles = filteredArticles[startIndex:]
	} else {
		filteredArticles = filteredArticles[startIndex:endIndex]
	}

	totalPages := (len(MOCK_ARTICLES) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"articles": filteredArticles,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      len(MOCK_ARTICLES),
			"totalPages": totalPages,
		},
	})
}

// GetArticleHandler returns a single article by slug
func GetArticleHandler(c *gin.Context) {
	slug := c.Param("slug")

	for _, article := range MOCK_ARTICLES {
		if article.Slug == slug {
			c.JSON(http.StatusOK, gin.H{"article": article})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
}

// GetArticleCategoriesHandler returns all article categories
func GetArticleCategoriesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"categories": MOCK_ARTICLE_CATEGORIES})
}

// GetAuthorsHandler returns all authors
func GetAuthorsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"authors": MOCK_AUTHORS})
}

// SetupArticlesRoutes sets up all articles routes
func SetupArticlesRoutes(router *gin.RouterGroup) {
	api := router.Group("/api/v1")
	{
		// Article endpoints
		api.GET("/articles", GetArticlesHandler)
		api.GET("/articles/:slug", GetArticleHandler)
		api.GET("/articles/categories", GetArticleCategoriesHandler)
		api.GET("/authors", GetAuthorsHandler)
	}
}
