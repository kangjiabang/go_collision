// internal/service/collision_service.go
package service

import (
	"collision_app_go/internal/repository"
	"collision_app_go/utils"
	"context"
)

type CollisionService struct {
	repo *repository.BuildingRepository
}

func NewCollisionService(repo *repository.BuildingRepository) *CollisionService {
	return &CollisionService{repo: repo}
}

// CheckCollision checks if a point collides with any buildings.
func (s *CollisionService) CheckCollision(ctx context.Context, longitude, latitude, height, collisionDistance float64) (map[string]interface{}, error) {
	utils.Infof("Checking collision for point: lon=%f, lat=%f, height=%f, distance=%f", longitude, latitude, height, collisionDistance)

	buildings, err := s.repo.GetCollisionBuildingsInfo(ctx, longitude, latitude, height, collisionDistance)
	if err != nil {
		return nil, err // Propagate error
	}

	isCollision := len(buildings) > 0

	response := map[string]interface{}{
		"status":       "success",
		"is_collision": isCollision,
	}

	if isCollision {
		response["building_infos"] = buildings
	}

	return response, nil
}
