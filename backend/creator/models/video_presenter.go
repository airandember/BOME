package models

import (
	"database/sql"
	"time"

	"bome-backend/infrastructure/database"
)

// VideoPresenter represents the many-to-many relationship between videos and presenters
type VideoPresenter struct {
	ID                     int            `json:"id"`
	VideoID                int            `json:"video_id"`
	PresenterID            int            `json:"presenter_id"`
	Role                   string         `json:"role"`
	AttributionPercentage  float64        `json:"attribution_percentage"`
	IsPrimary              bool           `json:"is_primary"`
	DisplayOrder           int            `json:"display_order"`
	AddedBy                sql.NullInt64  `json:"added_by,omitempty"`
	AddedAt                time.Time      `json:"added_at"`
	Notes                  sql.NullString `json:"notes,omitempty"`
}

// VideoPresenterWithDetails includes presenter and video details
type VideoPresenterWithDetails struct {
	ID                    int            `json:"id"`
	VideoID               int            `json:"video_id"`
	VideoTitle            string         `json:"video_title"`
	PresenterID           int            `json:"presenter_id"`
	PresenterName         string         `json:"presenter_name"`
	Role                  string         `json:"role"`
	AttributionPercentage float64        `json:"attribution_percentage"`
	IsPrimary             bool           `json:"is_primary"`
	DisplayOrder          int            `json:"display_order"`
	Notes                 sql.NullString `json:"notes,omitempty"`
	AddedAt               time.Time      `json:"added_at"`
}

// CreateVideoPresenterInput represents input for linking a presenter to a video
type CreateVideoPresenterInput struct {
	VideoID               int     `json:"video_id" binding:"required"`
	PresenterID           int     `json:"presenter_id" binding:"required"`
	Role                  string  `json:"role"`
	AttributionPercentage float64 `json:"attribution_percentage"`
	IsPrimary             bool    `json:"is_primary"`
	DisplayOrder          int     `json:"display_order"`
	AddedBy               int     `json:"added_by"`
	Notes                 string  `json:"notes"`
}

// UpdateVideoPresenterInput represents input for updating a video-presenter link
type UpdateVideoPresenterInput struct {
	Role                  *string  `json:"role"`
	AttributionPercentage *float64 `json:"attribution_percentage"`
	IsPrimary             *bool    `json:"is_primary"`
	DisplayOrder          *int     `json:"display_order"`
	Notes                 *string  `json:"notes"`
}

// CreateVideoPresenter links a presenter to a video
func CreateVideoPresenter(db *database.DB, input *CreateVideoPresenterInput) (*VideoPresenter, error) {
	vp := &VideoPresenter{}
	
	role := "presenter"
	if input.Role != "" {
		role = input.Role
	}
	
	attribution := 100.00
	if input.AttributionPercentage > 0 {
		attribution = input.AttributionPercentage
	}
	
	query := `
		INSERT INTO video_presenters (
			video_id, presenter_id, role, attribution_percentage,
			is_primary, display_order, added_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, video_id, presenter_id, role, attribution_percentage,
			is_primary, display_order, added_by, added_at, notes
	`
	
	err := db.QueryRow(
		query,
		input.VideoID, input.PresenterID, role, attribution,
		input.IsPrimary, input.DisplayOrder, input.AddedBy, input.Notes,
	).Scan(
		&vp.ID, &vp.VideoID, &vp.PresenterID, &vp.Role,
		&vp.AttributionPercentage, &vp.IsPrimary, &vp.DisplayOrder,
		&vp.AddedBy, &vp.AddedAt, &vp.Notes,
	)
	
	if err != nil {
		return nil, err
	}
	
	return vp, nil
}

// GetVideoPresentersByVideoID retrieves all presenters for a specific video
func GetVideoPresentersByVideoID(db *database.DB, videoID int) ([]*VideoPresenterWithDetails, error) {
	query := `
		SELECT 
			vp.id, vp.video_id, mvl.title as video_title,
			vp.presenter_id, p.name as presenter_name,
			vp.role, vp.attribution_percentage, vp.is_primary,
			vp.display_order, vp.notes, vp.added_at
		FROM video_presenters vp
		INNER JOIN presenters p ON vp.presenter_id = p.id
		INNER JOIN master_video_list mvl ON vp.video_id = mvl.id
		WHERE vp.video_id = $1
		ORDER BY vp.display_order ASC, vp.is_primary DESC
	`
	
	rows, err := db.Query(query, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	videoPresenters := []*VideoPresenterWithDetails{}
	
	for rows.Next() {
		vp := &VideoPresenterWithDetails{}
		err := rows.Scan(
			&vp.ID, &vp.VideoID, &vp.VideoTitle,
			&vp.PresenterID, &vp.PresenterName,
			&vp.Role, &vp.AttributionPercentage, &vp.IsPrimary,
			&vp.DisplayOrder, &vp.Notes, &vp.AddedAt,
		)
		if err != nil {
			return nil, err
		}
		videoPresenters = append(videoPresenters, vp)
	}
	
	return videoPresenters, nil
}

// GetVideoPresentersByPresenterID retrieves all videos for a specific presenter
func GetVideoPresentersByPresenterID(db *database.DB, presenterID int) ([]*VideoPresenterWithDetails, error) {
	query := `
		SELECT 
			vp.id, vp.video_id, mvl.title as video_title,
			vp.presenter_id, p.name as presenter_name,
			vp.role, vp.attribution_percentage, vp.is_primary,
			vp.display_order, vp.notes, vp.added_at
		FROM video_presenters vp
		INNER JOIN presenters p ON vp.presenter_id = p.id
		INNER JOIN master_video_list mvl ON vp.video_id = mvl.id
		WHERE vp.presenter_id = $1
		ORDER BY vp.added_at DESC
	`
	
	rows, err := db.Query(query, presenterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	videoPresenters := []*VideoPresenterWithDetails{}
	
	for rows.Next() {
		vp := &VideoPresenterWithDetails{}
		err := rows.Scan(
			&vp.ID, &vp.VideoID, &vp.VideoTitle,
			&vp.PresenterID, &vp.PresenterName,
			&vp.Role, &vp.AttributionPercentage, &vp.IsPrimary,
			&vp.DisplayOrder, &vp.Notes, &vp.AddedAt,
		)
		if err != nil {
			return nil, err
		}
		videoPresenters = append(videoPresenters, vp)
	}
	
	return videoPresenters, nil
}

// UpdateVideoPresenter updates a video-presenter link
func UpdateVideoPresenter(db *database.DB, id int, input *UpdateVideoPresenterInput) (*VideoPresenter, error) {
	// Build dynamic UPDATE query
	query := `UPDATE video_presenters SET`
	args := []interface{}{}
	argIndex := 1
	updates := []string{}
	
	if input.Role != nil {
		updates = append(updates, ` role = $`+string(rune('0'+argIndex)))
		args = append(args, *input.Role)
		argIndex++
	}
	if input.AttributionPercentage != nil {
		updates = append(updates, ` attribution_percentage = $`+string(rune('0'+argIndex)))
		args = append(args, *input.AttributionPercentage)
		argIndex++
	}
	if input.IsPrimary != nil {
		updates = append(updates, ` is_primary = $`+string(rune('0'+argIndex)))
		args = append(args, *input.IsPrimary)
		argIndex++
	}
	if input.DisplayOrder != nil {
		updates = append(updates, ` display_order = $`+string(rune('0'+argIndex)))
		args = append(args, *input.DisplayOrder)
		argIndex++
	}
	if input.Notes != nil {
		updates = append(updates, ` notes = $`+string(rune('0'+argIndex)))
		args = append(args, *input.Notes)
		argIndex++
	}
	
	if len(updates) == 0 {
		// Nothing to update, just return existing
		return GetVideoPresenterByID(db, id)
	}
	
	for i, update := range updates {
		if i > 0 {
			query += `,`
		}
		query += update
	}
	
	query += ` WHERE id = $` + string(rune('0'+argIndex))
	args = append(args, id)
	
	_, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	
	return GetVideoPresenterByID(db, id)
}

// GetVideoPresenterByID retrieves a specific video-presenter link by ID
func GetVideoPresenterByID(db *database.DB, id int) (*VideoPresenter, error) {
	vp := &VideoPresenter{}
	
	query := `
		SELECT id, video_id, presenter_id, role, attribution_percentage,
			is_primary, display_order, added_by, added_at, notes
		FROM video_presenters
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&vp.ID, &vp.VideoID, &vp.PresenterID, &vp.Role,
		&vp.AttributionPercentage, &vp.IsPrimary, &vp.DisplayOrder,
		&vp.AddedBy, &vp.AddedAt, &vp.Notes,
	)
	
	if err != nil {
		return nil, err
	}
	
	return vp, nil
}

// DeleteVideoPresenter removes a presenter from a video
func DeleteVideoPresenter(db *database.DB, id int) error {
	query := `DELETE FROM video_presenters WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// DeleteVideoPresentersByVideoID removes all presenters from a video
func DeleteVideoPresentersByVideoID(db *database.DB, videoID int) error {
	query := `DELETE FROM video_presenters WHERE video_id = $1`
	_, err := db.Exec(query, videoID)
	return err
}

// DeleteVideoPresentersByPresenterID removes all videos from a presenter
func DeleteVideoPresentersByPresenterID(db *database.DB, presenterID int) error {
	query := `DELETE FROM video_presenters WHERE presenter_id = $1`
	_, err := db.Exec(query, presenterID)
	return err
}

