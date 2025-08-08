// internal/repository/building_repo.go
package repository

import (
	"bufio"
	"collision_app_go/internal/model"
	"collision_app_go/utils"
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"

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

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		utils.Errorf("Failed to open file: %v", err)
		return map[string]interface{}{
			"success":  false,
			"code":     500,
			"errorMsg": "File not found",
		}, nil
	}
	defer file.Close()

	// Read all lines
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		utils.Errorf("Error reading file: %v", err)
		return map[string]interface{}{
			"success":  false,
			"code":     500,
			"errorMsg": fmt.Sprintf("Error reading file: %v", err),
		}, nil
	}

	utils.Infof("Read %d lines from file", len(lines))

	successCount := 0
	errorCount := 0

	// Process each line individually
	for lineNum, line := range lines {
		originalLineNum := lineNum + 1
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Process single record
		success, err := r.processSingleRecord(ctx, line, originalLineNum)
		if err != nil {
			utils.Infof("Line %d: Processing failed - %v", originalLineNum, err)
			errorCount++
			continue
		}

		if success {
			successCount++
			utils.Infof("✓ Line %d: Insert successful", originalLineNum)
		} else {
			errorCount++
			utils.Infof("✗ Line %d: Insert failed", originalLineNum)
		}
	}

	// Calculate success rate
	successRate := 0.0
	if len(lines) > 0 {
		successRate = float64(successCount) / float64(len(lines)) * 100
	}

	utils.Infof("File data insertion completed!")
	utils.Infof("Successfully inserted: %d", successCount)
	utils.Infof("Failed to process: %d", errorCount)
	utils.Infof("Overall success rate: %.1f%%", successRate)

	return map[string]interface{}{
		"success": true,
		"code":    200,
		"data": map[string]interface{}{
			"success_count": successCount,
			"error_count":   errorCount,
			"total_count":   len(lines),
			"success_rate":  successRate,
		},
	}, nil
}

func (r *BuildingRepository) processSingleRecord(ctx context.Context, line string, lineNum int) (bool, error) {
	// Parse the line (format: MULTIPOLYGON(((...))),13.56)
	parts := strings.Split(line, ",")
	if len(parts) < 2 {
		return false, fmt.Errorf("invalid format")
	}

	// Get WKT and height parts
	wktPart := strings.Join(parts[:len(parts)-1], ",")
	heightStr := parts[len(parts)-1]

	// Clean WKT and height strings
	wktGeom := strings.Trim(strings.TrimSpace(wktPart), `"'`)
	heightStr = strings.Trim(strings.TrimSpace(heightStr), `"'`)

	// Validate WKT format
	if !strings.HasPrefix(strings.ToUpper(wktGeom), "MULTIPOLYGON") || !strings.HasSuffix(wktGeom, "))") {
		return false, fmt.Errorf("invalid WKT format")
	}

	// Parse height
	buildingHeight, err := strconv.ParseFloat(heightStr, 64)
	if err != nil {
		return false, fmt.Errorf("invalid height format: %v", err)
	}

	// Generate building ID
	buildingID, err := GenerateBuildingIDPureCode(wktGeom)
	if err != nil {
		return false, fmt.Errorf("failed to generate building ID: %v", err)
	}

	// Execute single insert with individual transaction
	insertQuery := `
		INSERT INTO hzdk_buildings (geom, building_height, building_id)
		VALUES (ST_GeomFromText($1, 4326), $2, $3)
	`

	_, err = r.dbpool.Exec(ctx, insertQuery, wktGeom, buildingHeight, buildingID)
	if err != nil {
		return false, fmt.Errorf("database insert failed: %v", err)
	}

	return true, nil
}

func GenerateBuildingIDPureCode(wktGeom string) (int64, error) {
	// 使用哈希算法生成纯数字ID
	hash := fnv.New64a()
	hash.Write([]byte(wktGeom))
	id := int64(hash.Sum64())

	// 确保ID为正数
	if id < 0 {
		id = -id
	}

	// 如果ID为0，设置默认值
	if id == 0 {
		id = 1
	}

	return id, nil
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
