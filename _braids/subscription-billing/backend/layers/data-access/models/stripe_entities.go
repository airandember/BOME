package database

import (
	"database/sql"
	"encoding/json"
	"time"
)

// StripeEntity represents an abstraction layer for Stripe entities
type StripeEntity struct {
	ID         int
	EntityType string         // 'subscription', 'customer', 'payment_intent', 'price'
	EntityID   string         // Stripe ID
	LocalID    int            // Your internal ID
	Metadata   sql.NullString // JSON metadata
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CreateStripeEntity inserts a new Stripe entity mapping
func (db *DB) CreateStripeEntity(entityType, entityID string, localID int, metadata map[string]interface{}) (*StripeEntity, error) {
	var metadataJSON sql.NullString
	if metadata != nil {
		jsonData, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		metadataJSON.String = string(jsonData)
		metadataJSON.Valid = true
	}

	var id int
	err := db.QueryRow(
		`INSERT INTO stripe_entities (entity_type, entity_id, local_id, metadata, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id`,
		entityType, entityID, localID, metadataJSON,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetStripeEntityByID(id)
}

// GetStripeEntityByID retrieves a Stripe entity by ID
func (db *DB) GetStripeEntityByID(id int) (*StripeEntity, error) {
	entity := &StripeEntity{}
	err := db.QueryRow(
		`SELECT id, entity_type, entity_id, local_id, metadata, created_at, updated_at 
		 FROM stripe_entities WHERE id = $1`,
		id,
	).Scan(&entity.ID, &entity.EntityType, &entity.EntityID, &entity.LocalID,
		&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// GetStripeEntityByStripeID retrieves a Stripe entity by Stripe ID
func (db *DB) GetStripeEntityByStripeID(entityType, entityID string) (*StripeEntity, error) {
	entity := &StripeEntity{}
	err := db.QueryRow(
		`SELECT id, entity_type, entity_id, local_id, metadata, created_at, updated_at 
		 FROM stripe_entities WHERE entity_type = $1 AND entity_id = $2`,
		entityType, entityID,
	).Scan(&entity.ID, &entity.EntityType, &entity.EntityID, &entity.LocalID,
		&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// GetStripeEntityByLocalID retrieves a Stripe entity by local ID
func (db *DB) GetStripeEntityByLocalID(entityType string, localID int) (*StripeEntity, error) {
	entity := &StripeEntity{}
	err := db.QueryRow(
		`SELECT id, entity_type, entity_id, local_id, metadata, created_at, updated_at 
		 FROM stripe_entities WHERE entity_type = $1 AND local_id = $2`,
		entityType, localID,
	).Scan(&entity.ID, &entity.EntityType, &entity.EntityID, &entity.LocalID,
		&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// UpdateStripeEntityMetadata updates the metadata for a Stripe entity
func (db *DB) UpdateStripeEntityMetadata(id int, metadata map[string]interface{}) error {
	var metadataJSON sql.NullString
	if metadata != nil {
		jsonData, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON.String = string(jsonData)
		metadataJSON.Valid = true
	}

	_, err := db.Exec(`UPDATE stripe_entities SET metadata = $1, updated_at = NOW() WHERE id = $2`, metadataJSON, id)
	return err
}

// DeleteStripeEntity deletes a Stripe entity mapping
func (db *DB) DeleteStripeEntity(id int) error {
	_, err := db.Exec(`DELETE FROM stripe_entities WHERE id = $1`, id)
	return err
}

// GetStripeEntitiesByType retrieves all Stripe entities of a specific type
func (db *DB) GetStripeEntitiesByType(entityType string, limit, offset int) ([]*StripeEntity, error) {
	rows, err := db.Query(
		`SELECT id, entity_type, entity_id, local_id, metadata, created_at, updated_at 
		 FROM stripe_entities WHERE entity_type = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		entityType, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*StripeEntity
	for rows.Next() {
		entity := &StripeEntity{}
		err := rows.Scan(&entity.ID, &entity.EntityType, &entity.EntityID, &entity.LocalID,
			&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
