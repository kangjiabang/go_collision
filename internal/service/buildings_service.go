// internal/service/buildings_service.go
package service

import (
	"collision_app_go/utils"
	"context"
	"fmt"

	"collision_app_go/internal/repository"
)

type BuildingsService struct {
	repo *repository.BuildingRepository
}

func NewBuildingsService(repo *repository.BuildingRepository) *BuildingsService {
	return &BuildingsService{repo: repo}
}

// InsertBuildings handles the logic for inserting buildings from a file.
func (s *BuildingsService) InsertBuildings(ctx context.Context, filePath string) (map[string]interface{}, error) {
	utils.Infof("Service: Inserting buildings from file: %s", filePath)

	result, err := s.repo.InsertBuildingsFromFile(ctx, filePath)
	if err != nil {
		utils.Errorf("Service error during insert: %v", err)
		// Return a standardized error response
		return map[string]interface{}{
			"success":  false,
			"code":     501,
			"errorMsg": fmt.Sprintf("导入时发生异常: %v", err),
		}, nil // Return the error *response* as data
	}
	return result, nil
}

// UpdateBuildings handles the logic for updating all buildings.
func (s *BuildingsService) UpdateBuildings(ctx context.Context) (map[string]interface{}, error) {
	utils.Info("Service: Updating all buildings info...")

	result, err := s.repo.UpdateAllBuildingsInfoBatch(ctx)
	if err != nil {
		utils.Errorf("Service error during update: %v", err)
		// Return a standardized error response
		return map[string]interface{}{
			"success":  false,
			"code":     500,
			"errorMsg": fmt.Sprintf("更新时发生错误: %v", err),
		}, nil // Return the error *response* as data
	}
	return result, nil
}
