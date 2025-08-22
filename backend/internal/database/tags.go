package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Tag represents a tag in the system (updated to match actual DB schema)
type Tag struct {
	ID          int       `json:"id"`
	Word        string    `json:"word"`
	Frequency   int       `json:"frequency"`
	CategoryID  *int      `json:"category_id"` // Can be null
	SubsiteID   *int      `json:"subsite_id"`  // Can be null
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ActiveTag   bool      `json:"active_tag"`
	CategoryIDs []int     `json:"category_ids"`
}

// TagCategory represents a tag category (updated to match actual DB schema)
type TagCategory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	SubsiteIDs  []int     `json:"subsite_ids"` // Array of subsite IDs
	SubsiteID   *int      `json:"subsite_id"`  // Can be null
	UpdatedAt   time.Time `json:"updated_at"`
	TagIDs      []int     `json:"tag_ids"`
}

// GetTags retrieves all tags from the database
func (db *DB) GetTags() ([]Tag, error) {
	query := `
		SELECT id, word, frequency, category_id, subsite_id, created_at, updated_at, active_tag
		FROM tags
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
		var categoryID, subsiteID sql.NullInt64

		err := rows.Scan(
			&tag.ID,
			&tag.Word,
			&tag.Frequency,
			&categoryID,
			&subsiteID,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.ActiveTag,
		)
		if err != nil {
			log.Printf("Error scanning tag row: %v", err)
			continue
		}

		// Handle nullable fields
		if categoryID.Valid {
			id := int(categoryID.Int64)
			tag.CategoryID = &id
		}
		if subsiteID.Valid {
			id := int(subsiteID.Int64)
			tag.SubsiteID = &id
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// GetTagByID retrieves a specific tag by ID
func (db *DB) GetTagByID(tagID int) (*Tag, error) {
	query := `
		SELECT id, word, frequency, category_id, subsite_id, created_at, updated_at, active_tag
		FROM tags
		WHERE id = $1
	`

	var tag Tag
	var categoryID, subsiteID sql.NullInt64

	err := db.QueryRow(query, tagID).Scan(
		&tag.ID,
		&tag.Word,
		&tag.Frequency,
		&categoryID,
		&subsiteID,
		&tag.CreatedAt,
		&tag.UpdatedAt,
		&tag.ActiveTag,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to query tag: %v", err)
	}

	// Handle nullable fields
	if categoryID.Valid {
		id := int(categoryID.Int64)
		tag.CategoryID = &id
	}
	if subsiteID.Valid {
		id := int(subsiteID.Int64)
		tag.SubsiteID = &id
	}

	return &tag, nil
}

// CreateTag creates a new tag
func (db *DB) CreateTag(word string, subsiteID *int) (*Tag, error) {
	query := `
		INSERT INTO tags (word, frequency, category_id, subsite_id, active_tag, created_at, updated_at)
		VALUES ($1, 1, NULL, $2, true, NOW(), NOW())
		RETURNING id, word, frequency, category_id, subsite_id, created_at, updated_at, active_tag
	`

	var tag Tag
	var categoryID, subsiteIDScan sql.NullInt64

	err := db.QueryRow(query, word, subsiteID).Scan(
		&tag.ID,
		&tag.Word,
		&tag.Frequency,
		&categoryID,
		&subsiteIDScan,
		&tag.CreatedAt,
		&tag.UpdatedAt,
		&tag.ActiveTag,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %v", err)
	}

	// Handle nullable fields
	if categoryID.Valid {
		id := int(categoryID.Int64)
		tag.CategoryID = &id
	}
	if subsiteIDScan.Valid {
		id := int(subsiteIDScan.Int64)
		tag.SubsiteID = &id
	}

	return &tag, nil
}

// UpdateTag updates an existing tag
func (db *DB) UpdateTag(tagID int, word string) error {
	query := `
		UPDATE tags 
		SET word = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := db.Exec(query, tagID, word)
	if err != nil {
		return fmt.Errorf("failed to update tag: %v", err)
	}

	return nil
}

// DeleteTag deletes a tag
func (db *DB) DeleteTag(tagID int) error {
	query := `DELETE FROM tags WHERE id = $1`

	_, err := db.Exec(query, tagID)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %v", err)
	}

	return nil
}

// UpdateTagCategory updates the category_id for a tag
func (db *DB) UpdateTagCategory(tagID int, categoryID *int) error {
	query := `
		UPDATE tags 
		SET category_id = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := db.Exec(query, tagID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to update tag category: %v", err)
	}

	return nil
}

// GetTagCategoryByID retrieves a specific tag category by ID
func (db *DB) GetTagCategoryByID(categoryID int) (*TagCategory, error) {
	query := `
		SELECT id, name, description, color, created_at, subsite_ids, subsite_id, updated_at
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
		&category.SubsiteIDs,
		&subsiteID,
		&category.UpdatedAt,
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
		INSERT INTO tag_categories (name, color, description, subsite_ids, subsite_id, created_at, updated_at)
		VALUES ($1, $2, $3, ARRAY[$4], $4, NOW(), NOW())
		RETURNING id, name, description, color, created_at, subsite_ids, subsite_id, updated_at
	`

	var category TagCategory
	var subsiteIDScan sql.NullInt64

	err := db.QueryRow(query, name, color, description, subsiteID).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Color,
		&category.CreatedAt,
		&category.SubsiteIDs,
		&subsiteIDScan,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag category: %v", err)
	}

	// Handle nullable subsite_id
	if subsiteIDScan.Valid {
		id := int(subsiteIDScan.Int64)
		category.SubsiteID = &id
	}

	return &category, nil
}

// UpdateTagCategoryFields updates an existing tag category
func (db *DB) UpdateTagCategoryFields(categoryID int, name, color, description string) error {
	query := `
		UPDATE tag_categories 
		SET name = $2, color = $3, description = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := db.Exec(query, categoryID, name, color, description)
	if err != nil {
		return fmt.Errorf("failed to update tag category: %v", err)
	}

	return nil
}

// DeleteTagCategory deletes a tag category
func (db *DB) DeleteTagCategory(categoryID int) error {
	query := `DELETE FROM tag_categories WHERE id = $1`

	_, err := db.Exec(query, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete tag category: %v", err)
	}

	return nil
}

// GetCategoryTags retrieves all tags for a specific category
func (db *DB) GetCategoryTags(categoryID int) ([]Tag, error) {
	query := `
		SELECT id, word, frequency, category_id, subsite_id, created_at, updated_at, active_tag
		FROM tags
		WHERE category_id = $1
		ORDER BY word ASC
	`

	rows, err := db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query category tags: %v", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var categoryID, subsiteID sql.NullInt64

		err := rows.Scan(
			&tag.ID,
			&tag.Word,
			&tag.Frequency,
			&categoryID,
			&subsiteID,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.ActiveTag,
		)
		if err != nil {
			log.Printf("Error scanning category tag row: %v", err)
			continue
		}

		// Handle nullable fields
		if categoryID.Valid {
			id := int(categoryID.Int64)
			tag.CategoryID = &id
		}
		if subsiteID.Valid {
			id := int(subsiteID.Int64)
			tag.SubsiteID = &id
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// BatchUpdateTagCategories handles batch updates for tag-category relationships
func (db *DB) BatchUpdateTagCategories(changes []map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	for _, change := range changes {
		tagID := int(change["tagId"].(float64))
		categoryID := change["categoryId"]
		action := change["action"].(string)

		if action == "add" {
			// Add tag to category
			if categoryID != nil {
				catID := int(categoryID.(float64))
				err = addTagToCategory(tx, tagID, catID)
				if err != nil {
					return fmt.Errorf("failed to add tag %d to category %d: %v", tagID, catID, err)
				}
			}
		} else if action == "remove" {
			// Remove tag from category
			err = removeTagFromCategory(tx, tagID)
			if err != nil {
				return fmt.Errorf("failed to remove tag %d from category: %v", tagID, err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// Helper function to add tag to category
func addTagToCategory(tx *sql.Tx, tagID, categoryID int) error {
	// Update tag's category_id
	query := `
		UPDATE tags 
		SET category_id = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := tx.Exec(query, tagID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to update tag category_id: %v", err)
	}

	return nil
}

// Helper function to remove tag from category
func removeTagFromCategory(tx *sql.Tx, tagID int) error {
	// Clear tag's category_id
	query := `
		UPDATE tags 
		SET category_id = NULL, updated_at = NOW()
		WHERE id = $1
	`
	_, err := tx.Exec(query, tagID)
	if err != nil {
		return fmt.Errorf("failed to clear tag category_id: %v", err)
	}

	return nil
}

// GetTagsBySubsite retrieves tags for a specific subsite
func (db *DB) GetTagsBySubsite(subsiteID int) ([]Tag, error) {
	query := `
		SELECT id, word, frequency, category_id, subsite_id, created_at, updated_at, active_tag
		FROM tags
		WHERE subsite_id = $1
		ORDER BY word ASC
	`

	rows, err := db.Query(query, subsiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subsite tags: %v", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var categoryID, subsiteIDScan sql.NullInt64

		err := rows.Scan(
			&tag.ID,
			&tag.Word,
			&tag.Frequency,
			&categoryID,
			&subsiteIDScan,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.ActiveTag,
		)
		if err != nil {
			log.Printf("Error scanning subsite tag row: %v", err)
			continue
		}

		// Handle nullable fields
		if categoryID.Valid {
			id := int(categoryID.Int64)
			tag.CategoryID = &id
		}
		if subsiteIDScan.Valid {
			id := int(subsiteIDScan.Int64)
			tag.SubsiteID = &id
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// GetTagCategoriesBySubsite retrieves tag categories for a specific subsite
func (db *DB) GetTagCategoriesBySubsite(subsiteID int) ([]TagCategory, error) {
	query := `
		SELECT id, name, description, color, created_at, subsite_ids, subsite_id, updated_at
		FROM tag_categories
		WHERE subsite_id = $1 OR $1 = ANY(subsite_ids)
		ORDER BY name ASC
	`

	rows, err := db.Query(query, subsiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subsite tag categories: %v", err)
	}
	defer rows.Close()

	var categories []TagCategory
	for rows.Next() {
		var category TagCategory
		var subsiteIDScan sql.NullInt64

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Color,
			&category.CreatedAt,
			&category.SubsiteIDs,
			&subsiteIDScan,
			&category.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning subsite tag category row: %v", err)
			continue
		}

		// Handle nullable subsite_id
		if subsiteIDScan.Valid {
			id := int(subsiteIDScan.Int64)
			category.SubsiteID = &id
		}

		categories = append(categories, category)
	}

	return categories, nil
}
