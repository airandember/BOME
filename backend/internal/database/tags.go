package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

// Tag represents a tag in the system with bidirectional category relationships
type Tag struct {
	ID              int       `json:"id"`
	Word            string    `json:"word"`
	Frequency       int       `json:"frequency"`
	SubsiteIDOrigin *int      `json:"subsite_id_origin"` // Which subsite created this tag (immutable)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ActiveTag       bool      `json:"active_tag"`
	CategoryIDs     []int     `json:"category_ids"` // Array field for many-to-many relationships
	SubsiteIDs      []int     `json:"subsite_ids"`  // Which subsites use this tag (mutable array)
}

// TagCategory represents a tag category with bidirectional tag relationships
type TagCategory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	SubsiteIDs  []int     `json:"subsite_ids"` // Array of subsite IDs
	SubsiteID   *int      `json:"subsite_id"`  // Legacy single subsite (keep for compatibility)
	UpdatedAt   time.Time `json:"updated_at"`
	TagIDs      []int     `json:"tag_ids"` // Array field for many-to-many relationships
}

// ===== TAG CRUD OPERATIONS =====

// GetTags retrieves all tags with their category relationships for streaming subsite
func (db *DB) GetTags() ([]Tag, error) {
	query := `
		SELECT id, word, frequency, subsite_id_origin,  created_at, 
		       updated_at, active_tag, COALESCE(category_ids, '{}'), COALESCE(subsite_ids, '{}')
		FROM tags
		WHERE subsite_id_origin = 1 OR 1 = ANY(COALESCE(subsite_ids, '{}'))
		ORDER BY word ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %v", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var subsiteIDOrigin sql.NullInt64

		err := rows.Scan(
			&tag.ID,
			&tag.Word,
			&tag.Frequency,
			&subsiteIDOrigin,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.ActiveTag,
			pq.Array(&tag.CategoryIDs),
			pq.Array(&tag.SubsiteIDs),
		)
		if err != nil {
			log.Printf("❌ Error scanning tag row: %v", err)
			continue
		}

		// Handle nullable subsite_id_origin
		if subsiteIDOrigin.Valid {
			id := int(subsiteIDOrigin.Int64)
			tag.SubsiteIDOrigin = &id
		}

		tags = append(tags, tag)
	}

	log.Printf("✅ Retrieved %d tags for streaming subsite (ID: 1)", len(tags))
	return tags, nil
}

// GetTagByID retrieves a specific tag by ID
func (db *DB) GetTagByID(tagID int) (*Tag, error) {
	query := `
		SELECT id, word, frequency, subsite_id_origin, created_at, updated_at, active_tag, 
       COALESCE(category_ids, '{}'), COALESCE(subsite_ids, '{}')
		FROM tags
		WHERE id = $1
	`

	var tag Tag
	var subsiteIDOrigin sql.NullInt64
	var subsiteIDs pq.Int64Array // Use pq.Int64Array instead

	err := db.QueryRow(query, tagID).Scan(
		&tag.ID,
		&tag.Word,
		&tag.Frequency,
		&subsiteIDOrigin,
		&subsiteIDs, // Scan into pq.Int64Array
		&tag.CreatedAt,
		&tag.UpdatedAt,
		&tag.ActiveTag,
		pq.Array(&tag.CategoryIDs),
		pq.Array(&tag.SubsiteIDs),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to query tag: %v", err)
	}

	// Handle nullable subsite_id_origin
	if subsiteIDOrigin.Valid {
		id := int(subsiteIDOrigin.Int64)
		tag.SubsiteIDOrigin = &id
	}

	// Convert pq.Int64Array to []int
	tag.SubsiteIDs = make([]int, len(subsiteIDs))
	for i, v := range subsiteIDs {
		tag.SubsiteIDs[i] = int(v)
	}

	return &tag, nil
}

// ===== TAG CATEGORY CRUD OPERATIONS =====

// TagGetCategories retrieves all tag categories for streaming subsite
func (db *DB) TagGetCategories() ([]TagCategory, error) {
	query := `
		SELECT id, name, description, color, created_at, 
		       COALESCE(subsite_ids, '{}'), subsite_id, updated_at, COALESCE(tag_ids, '{}')
		FROM tag_categories
		WHERE subsite_id = 1 OR 1 = ANY(COALESCE(subsite_ids, '{}'))
		ORDER BY name ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag categories: %v", err)
	}
	defer rows.Close()

	var categories []TagCategory
	for rows.Next() {
		var category TagCategory
		var subsiteID sql.NullInt64

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Color,
			&category.CreatedAt,
			pq.Array(&category.SubsiteIDs), // Change from &subsiteIDs to pq.Array(&category.SubsiteIDs)
			&subsiteID,
			&category.UpdatedAt,
			pq.Array(&category.TagIDs), // Change from &tagIDs to pq.Array(&category.TagIDs)
		)
		if err != nil {
			log.Printf("❌ TagGetCategories: Error scanning tag_categories row: %v", err)
			continue
		}

		// Handle nullable subsite_id
		if subsiteID.Valid {
			id := int(subsiteID.Int64)
			category.SubsiteID = &id
		}

		categories = append(categories, category)
	}

	log.Printf("✅ Retrieved %d tag categories for streaming subsite (ID: 1)", len(categories))
	return categories, nil
}

// GetTagCategoryByID retrieves a specific tag category by ID
func (db *DB) GetTagCategoryByID(categoryID int) (*TagCategory, error) {
	query := `
		SELECT id, name, description, color, created_at, 
		       COALESCE(subsite_ids, '{}'), subsite_id, updated_at, COALESCE(tag_ids, '{}')
		FROM tag_categories
		WHERE id = $1
	`

	var category TagCategory
	var subsiteID sql.NullInt64

	err := db.QueryRow(query, categoryID).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Color,
		&category.CreatedAt,
		pq.Array(&category.SubsiteIDs),
		&subsiteID,
		&category.UpdatedAt,
		pq.Array(&category.TagIDs),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag category not found")
		}
		return nil, fmt.Errorf("failed to query tag category: %v", err)
	}

	// Handle nullable subsite_id
	if subsiteID.Valid {
		id := int(subsiteID.Int64)
		category.SubsiteID = &id
	}

	return &category, nil
}

// CreateTagCategory creates a new tag category
func (db *DB) CreateTagCategory(name, color, description string, subsiteID *int) (*TagCategory, error) {
	query := `
		INSERT INTO tag_categories (name, color, description, subsite_ids, subsite_id, created_at, updated_at, tag_ids)
		VALUES ($1, $2, $3, COALESCE($4::int[], '{}'), $4, NOW(), NOW(), '{}')
		RETURNING id, name, description, color, created_at, 
		          COALESCE(subsite_ids, '{}'), subsite_id, updated_at, COALESCE(tag_ids, '{}')
	`

	var category TagCategory
	var subsiteIDScan sql.NullInt64
	var subsiteIDs []int
	if subsiteID != nil {
		subsiteIDs = []int{*subsiteID}
	}

	err := db.QueryRow(query, name, color, description, pq.Array(subsiteIDs)).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Color,
		&category.CreatedAt,
		pq.Array(&category.SubsiteIDs),
		&subsiteIDScan,
		&category.UpdatedAt,
		pq.Array(&category.TagIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag category: %v", err)
	}

	// Handle nullable subsite_id
	if subsiteIDScan.Valid {
		id := int(subsiteIDScan.Int64)
		category.SubsiteID = &id
	}

	log.Printf("✅ Created tag category %d: %s", category.ID, category.Name)
	return &category, nil
}

// UpdateTagCategoryFields updates an existing tag category's basic fields
func (db *DB) UpdateTagCategoryFields(categoryID int, name, color, description string) error {
	query := `
		UPDATE tag_categories 
		SET name = $2, color = $3, description = $4, updated_at = NOW()
		WHERE id = $1
	`

	result, err := db.Exec(query, categoryID, name, color, description)
	if err != nil {
		return fmt.Errorf("failed to update tag category: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag category not found")
	}

	log.Printf("✅ Updated tag category %d", categoryID)
	return nil
}

// DeleteTagCategory deletes a tag category and removes it from all tags
func (db *DB) DeleteTagCategory(categoryID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Remove category from all tags' category_ids arrays
	_, err = tx.Exec(`
		UPDATE tags 
		SET category_ids = array_remove(category_ids, $1), updated_at = NOW()
		WHERE $1 = ANY(category_ids)
	`, categoryID)
	if err != nil {
		return fmt.Errorf("failed to remove category from tags: %v", err)
	}

	// Delete the category
	result, err := tx.Exec(`DELETE FROM tag_categories WHERE id = $1`, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete tag category: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag category not found")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ Deleted tag category %d", categoryID)
	return nil
}

// ===== BIDIRECTIONAL RELATIONSHIP OPERATIONS =====

// AddTagToCategory adds a tag to a category (bidirectional update)
func (db *DB) AddTagToCategory(tagID, categoryID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	log.Printf("🔗 Adding tag %d to category %d", tagID, categoryID)

	// Add tag to category's tag_ids array (if not already present)
	_, err = tx.Exec(`
		UPDATE tag_categories 
		SET tag_ids = array_append(tag_ids, $1), updated_at = NOW()
		WHERE id = $2 AND NOT ($1 = ANY(tag_ids))
	`, tagID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to add tag to category tag_ids: %v", err)
	}

	// Add category to tag's category_ids array (if not already present)
	_, err = tx.Exec(`
		UPDATE tags 
		SET category_ids = array_append(category_ids, $1), updated_at = NOW()
		WHERE id = $2 AND NOT ($1 = ANY(category_ids))
	`, categoryID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add category to tag category_ids: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ Successfully added tag %d to category %d", tagID, categoryID)
	return nil
}

// TagRemoveFromCategory removes a tag from a specific category (bidirectional update)
func (db *DB) TagRemoveFromCategory(tagID, categoryID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	log.Printf("🔗 Removing tag %d from category %d", tagID, categoryID)

	// Remove tag from category's tag_ids array
	_, err = tx.Exec(`
		UPDATE tag_categories 
		SET tag_ids = array_remove(tag_ids, $1), updated_at = NOW()
		WHERE id = $2
	`, tagID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from category tag_ids: %v", err)
	}

	// Remove category from tag's category_ids array
	_, err = tx.Exec(`
		UPDATE tags 
		SET category_ids = array_remove(category_ids, $1), updated_at = NOW()
		WHERE id = $2
	`, categoryID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove category from tag category_ids: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ Successfully removed tag %d from category %d", tagID, categoryID)
	return nil
}

// RemoveTagFromAllCategories removes a tag from all categories
func (db *DB) RemoveTagFromAllCategories(tagID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	log.Printf("🔗 Removing tag %d from all categories", tagID)

	// Remove tag from all categories' tag_ids arrays
	_, err = tx.Exec(`
		UPDATE tag_categories 
		SET tag_ids = array_remove(tag_ids, $1), updated_at = NOW()
		WHERE $1 = ANY(tag_ids)
	`, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from all categories: %v", err)
	}

	// Clear tag's category_ids array
	_, err = tx.Exec(`
		UPDATE tags 
		SET category_ids = '{}', updated_at = NOW()
		WHERE id = $1
	`, tagID)
	if err != nil {
		return fmt.Errorf("failed to clear tag category_ids: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ Successfully removed tag %d from all categories", tagID)
	return nil
}

// ===== QUERY OPERATIONS =====

// GetTagsByCategory retrieves all tags for a specific category
func (db *DB) GetTagsByCategory(categoryID int) ([]Tag, error) {
	query := `
		SELECT id, word, frequency, subsite_id_origin, COALESCE(subsite_ids, '{}'), created_at, 
		       updated_at, active_tag, COALESCE(category_ids, '{}')
		FROM tags
		WHERE $1 = ANY(category_ids) AND active_tag = true
		ORDER BY word ASC
	`

	rows, err := db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags for category %d: %v", categoryID, err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var subsiteIDOrigin sql.NullInt64
		var subsiteIDs pq.Int64Array // Use pq.Int64Array instead

		err := rows.Scan(
			&tag.ID,
			&tag.Word,
			&tag.Frequency,
			&subsiteIDOrigin,
			&subsiteIDs, // Scan into pq.Int64Array
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.ActiveTag,
			pq.Array(&tag.CategoryIDs),
			pq.Array(&tag.SubsiteIDs),
		)
		if err != nil {
			log.Printf("❌  GetTagsByCategory: Error scanning tag row: %v", err)
			continue
		}

		// Handle nullable subsite_id_origin
		if subsiteIDOrigin.Valid {
			id := int(subsiteIDOrigin.Int64)
			tag.SubsiteIDOrigin = &id
		}

		// Convert pq.Int64Array to []int
		tag.SubsiteIDs = make([]int, len(subsiteIDs))
		for i, v := range subsiteIDs {
			tag.SubsiteIDs[i] = int(v)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// GetCategoriesForTag retrieves all categories for a specific tag
func (db *DB) GetCategoriesForTag(tagID int) ([]TagCategory, error) {
	query := `
		SELECT id, name, description, color, created_at, 
		       COALESCE(subsite_ids, '{}'), subsite_id, updated_at, COALESCE(tag_ids, '{}')
		FROM tag_categories
		WHERE $1 = ANY(tag_ids)
		ORDER BY name ASC
	`

	rows, err := db.Query(query, tagID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories for tag %d: %v", tagID, err)
	}
	defer rows.Close()

	var categories []TagCategory
	for rows.Next() {
		var category TagCategory
		var subsiteID sql.NullInt64

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Color,
			&category.CreatedAt,

			&subsiteID,
			&category.UpdatedAt,
			pq.Array(&category.TagIDs),
			pq.Array(&category.SubsiteIDs),
		)
		if err != nil {
			log.Printf("❌  GetCategoriesForTag: Error scanning category row: %v", err)
			continue
		}

		// Handle nullable subsite_id
		if subsiteID.Valid {
			id := int(subsiteID.Int64)
			category.SubsiteID = &id
		}

		categories = append(categories, category)
	}

	return categories, nil
}

// ===== LEGACY COMPATIBILITY FUNCTIONS =====

// These functions maintain compatibility with existing code while using the new array system

// CreateTag creates a new tag
func (db *DB) CreateTag(word string, subsiteID *int) (*Tag, error) {
	var subsiteIDs []int
	if subsiteID != nil {
		subsiteIDs = []int{*subsiteID}
	}

	query := `
		INSERT INTO tags (word, frequency, subsite_id_origin, subsite_ids, active_tag, created_at, updated_at, category_ids)
		VALUES ($1, 1, $2, $3, true, NOW(), NOW(), '{}')
		RETURNING id, word, frequency, subsite_id_origin, COALESCE(subsite_ids, '{}'), created_at, updated_at, active_tag, COALESCE(category_ids, '{}')
	`

	var tag Tag
	var subsiteIDOriginScan sql.NullInt64
	var subsiteIDsArray pq.Int64Array

	err := db.QueryRow(query, word, subsiteID, pq.Array(subsiteIDs)).Scan(
		&tag.ID,
		&tag.Word,
		&tag.Frequency,
		&subsiteIDOriginScan,
		&subsiteIDsArray,
		&tag.CreatedAt,
		&tag.UpdatedAt,
		&tag.ActiveTag,
		pq.Array(&tag.CategoryIDs),
		pq.Array(&tag.SubsiteIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %v", err)
	}

	// Handle nullable subsite_id_origin
	if subsiteIDOriginScan.Valid {
		id := int(subsiteIDOriginScan.Int64)
		tag.SubsiteIDOrigin = &id
	}

	// Convert array
	tag.SubsiteIDs = make([]int, len(subsiteIDsArray))
	for i, v := range subsiteIDsArray {
		tag.SubsiteIDs[i] = int(v)
	}

	log.Printf("✅ Created tag %d: %s", tag.ID, tag.Word)
	return &tag, nil
}

// UpdateTag updates an existing tag (legacy compatibility)
func (db *DB) UpdateTag(tagID int, word string) error {
	query := `
		UPDATE tags 
		SET word = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := db.Exec(query, tagID, word)
	if err != nil {
		return fmt.Errorf("failed to update tag: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}

	log.Printf("✅ Updated tag %d", tagID)
	return nil
}

// DeleteTag deletes a tag (legacy compatibility)
func (db *DB) DeleteTag(tagID int) error {
	// First remove from all categories
	err := db.RemoveTagFromAllCategories(tagID)
	if err != nil {
		log.Printf("⚠️ Warning: failed to remove tag %d from categories: %v", tagID, err)
	}

	// Then delete the tag
	query := `DELETE FROM tags WHERE id = $1`
	result, err := db.Exec(query, tagID)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}

	log.Printf("✅ Deleted tag %d", tagID)
	return nil
}

// GetCategoryTags is an alias for GetTagsByCategory (legacy compatibility)
func (db *DB) GetCategoryTags(categoryID int) ([]Tag, error) {
	return db.GetTagsByCategory(categoryID)
}

// GetTagsBySubsite retrieves all tags for a specific subsite
func (db *DB) GetTagsBySubsite(subsiteID int) ([]Tag, error) {
	query := `
		SELECT id, word, frequency, subsite_id_origin, created_at, 
		       updated_at, active_tag, COALESCE(category_ids, '{}'), COALESCE(subsite_ids, '{}')
		FROM tags
		WHERE active_tag = true AND (subsite_id_origin = $1 OR $1 = ANY(COALESCE(subsite_ids, '{}')))
		ORDER BY word ASC
	`

	rows, err := db.Query(query, subsiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags for subsite %d: %v", subsiteID, err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var subsiteIDOrigin sql.NullInt64

		err := rows.Scan(
			&tag.ID,
			&tag.Word,
			&tag.Frequency,
			&subsiteIDOrigin,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.ActiveTag,
			pq.Array(&tag.CategoryIDs),
			pq.Array(&tag.SubsiteIDs), // ADD THIS LINE
		)
		if err != nil {
			log.Printf("❌  GetTagsBySubsite: Error scanning tag row: %v", err)
			continue
		}

		// Handle nullable subsite_id_origin
		if subsiteIDOrigin.Valid {
			id := int(subsiteIDOrigin.Int64)
			tag.SubsiteIDOrigin = &id
		}

		// Convert pq.Int64Array to []int
		// tag.SubsiteIDs = make([]int, len(subsiteIDs))
		// for i, v := range subsiteIDs {
		// 	tag.SubsiteIDs[i] = int(v)
		// }

		tags = append(tags, tag)
	}

	log.Printf("✅ Retrieved %d tags for subsite %d", len(tags), subsiteID)
	return tags, nil
}

// Fix GetTagCategoriesBySubsite function (lines 745-759)
func (db *DB) GetTagCategoriesBySubsite(subsiteID int) ([]TagCategory, error) {
	query := `
		SELECT id, name, description, color, created_at, COALESCE(subsite_ids, '{}'), subsite_id, updated_at, COALESCE(tag_ids, '{}')
		FROM tag_categories
		WHERE subsite_id = $1 OR $1 = ANY(COALESCE(subsite_ids, '{}'))
		ORDER BY name ASC
	`

	rows, err := db.Query(query, subsiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag categories for subsite %d: %v", subsiteID, err)
	}
	defer rows.Close()

	var categories []TagCategory
	for rows.Next() {
		var category TagCategory
		var subsiteIDScan sql.NullInt64
		var subsiteIDs pq.Int64Array // Add this
		var tagIDs pq.Int64Array     // Add this

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Color,
			&category.CreatedAt,
			&subsiteIDs, // Change from pq.Array(&category.SubsiteIDs)
			&subsiteIDScan,
			&category.UpdatedAt,
			&tagIDs, // Change from pq.Array(&category.TagIDs)
		)
		if err != nil {
			log.Printf("❌  GetTagCategoriesBySubsite: Error scanning tag_category table row: %v", err)
			continue
		}

		// Convert pq.Int64Array to []int (ADD THIS)
		category.SubsiteIDs = make([]int, len(subsiteIDs))
		for i, v := range subsiteIDs {
			category.SubsiteIDs[i] = int(v)
		}

		category.TagIDs = make([]int, len(tagIDs))
		for i, v := range tagIDs {
			category.TagIDs[i] = int(v)
		}

		// Handle nullable subsite_id
		if subsiteIDScan.Valid {
			id := int(subsiteIDScan.Int64)
			category.SubsiteID = &id
		}

		categories = append(categories, category)
	}

	log.Printf("✅ Retrieved %d tag categories for subsite %d", len(categories), subsiteID)
	return categories, nil
}
