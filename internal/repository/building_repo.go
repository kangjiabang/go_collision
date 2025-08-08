// internal/repository/building_repo.go
package repository

import (
	"collision_app_go/internal/model"
	"collision_app_go/utils"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BuildingRepository struct {
	dbpool *pgxpool.Pool
}

func NewBuildingRepository(dbpool *pgxpool.Pool) *BuildingRepository {
	return &BuildingRepository{dbpool: dbpool}
}

// GetCollisionBuildingsInfo finds buildings colliding with a point.
func (r *BuildingRepository) GetCollisionBuildingsInfo(ctx context.Context, longitude, latitude, height, collisionDistance float64) ([]model.Building, error) {
	query := `
        SELECT 
            building_id, building_name, ST_AsText(geom) AS geom, building_height
        FROM 
            hzdk_buildings
        WHERE 
            ST_DWithin(
                geom::geography, 
                ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
                $3)
            AND $4 < building_height
    `

	utils.Debug("Executing SQL: %s with args: lon=%f, lat=%f, dist=%f, height=%f", query, longitude, latitude, collisionDistance, height)

	rows, err := r.dbpool.Query(ctx, query, longitude, latitude, collisionDistance, height)
	if err != nil {
		utils.Errorf("Database query failed: %v", err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var buildings []model.Building
	for rows.Next() {
		var b model.Building
		err := rows.Scan(&b.BuildingID, &b.BuildingName, &b.Geom, &b.BuildingHeight)
		if err != nil {
			utils.Errorf("Failed to scan row: %v", err)
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		buildings = append(buildings, b)
	}

	if err = rows.Err(); err != nil {
		utils.Errorf("Row iteration error: %v", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return buildings, nil
}

// Placeholder for insert logic - you need to implement file reading/parsing
func (r *BuildingRepository) InsertBuildingsFromFile(ctx context.Context, filePath string) (map[string]interface{}, error) {
	utils.Infof("Inserting buildings from file: %s", filePath)
	// TODO: Implement actual file reading and database insertion logic
	// Example:
	// 1. Open file at filePath
	// 2. Parse data (CSV?)
	// 3. Prepare INSERT statements or use pgx.CopyFrom
	// 4. Execute inserts
	// 5. Handle errors
	// For now, simulate success
	return map[string]interface{}{
		"success": true,
		"code":    200,
		"message": fmt.Sprintf("Buildings inserted from %s (placeholder)", filePath),
	}, nil
}

// Placeholder for update logic
func (r *BuildingRepository) UpdateAllBuildingsInfoBatch(ctx context.Context) (map[string]interface{}, error) {
	utils.Info("Updating all buildings info (batch)...")
	// TODO: Implement actual update logic
	// Example:
	// 1. Query for buildings needing update
	// 2. Process them
	// 3. Update records
	// For now, simulate success
	return map[string]interface{}{
		"success": true,
		"code":    200,
		"message": "All buildings info updated (placeholder)",
	}, nil
}
