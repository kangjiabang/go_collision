// internal/handler/handler.go
package handler

import (
	"collision_app_go/utils"
	"fmt"
	"net/http"
	"strconv"

	"collision_app_go/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	collisionService *service.CollisionService
	buildingsService *service.BuildingsService
}

func NewHandler(collisionService *service.CollisionService, buildingsService *service.BuildingsService) *Handler {
	return &Handler{
		collisionService: collisionService,
		buildingsService: buildingsService,
	}
}

// CollisionInfo godoc
func (h *Handler) CollisionInfo(c *gin.Context) {
	longitudeStr := c.Query("longitude")
	latitudeStr := c.Query("latitude")
	heightStr := c.Query("height")
	collisionDistanceStr := c.DefaultQuery("collision_distance", "2")

	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid longitude"})
		return
	}
	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid latitude"})
		return
	}
	height, err := strconv.ParseFloat(heightStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid height"})
		return
	}
	collisionDistance, err := strconv.ParseFloat(collisionDistanceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid collision_distance"})
		return
	}

	utils.Infof("Received request: longitude=%f, latitude=%f, height=%f, collision_distance=%f", longitude, latitude, height, collisionDistance)

	result, err := h.collisionService.CheckCollision(c.Request.Context(), longitude, latitude, height, collisionDistance)
	if err != nil {
		utils.Errorf("Service error in CollisionInfo: %v", err)
		// 如果 Service 返回错误，直接返回 500 和错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("检测碰撞时发生错误: %v", err)})
		return
	}

	// 如果 Service 成功，result 就是期望的 map[string]interface{}
	c.JSON(http.StatusOK, result)
}

// InsertBuildingsInfo godoc
func (h *Handler) InsertBuildingsInfo(c *gin.Context) {
	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": 400, "errorMsg": "文件路径不能为空"})
		return
	}

	utils.Infof("Received request to insert buildings from file: %s", filePath)

	// Service 负责处理文件读取和数据库插入，并返回结果 map 或 error
	result, err := h.buildingsService.InsertBuildings(c.Request.Context(), filePath)
	if err != nil {
		utils.Errorf("Service error in InsertBuildingsInfo: %v", err)
		// 如果 Service 返回错误，直接返回 500 和错误信息
		// 注意：这里假设 Service 在出错时返回 error，而不是一个错误的 map
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":  false,
			"code":     500, // 或者从 err 中解析出更具体的错误码
			"errorMsg": fmt.Sprintf("导入建筑物信息发生错误: %v", err),
		})
		return
	}

	// 如果 Service 成功，result 就是期望的 map[string]interface{} (例如 {"success": true, ...})
	c.JSON(http.StatusOK, result)
}

// UpdateBuildingsInfo godoc
func (h *Handler) UpdateBuildingsInfo(c *gin.Context) {
	utils.Info("Received request to update all buildings info...")

	// Service 负责处理更新逻辑，并返回结果 map 或 error
	result, err := h.buildingsService.UpdateBuildings(c.Request.Context())
	if err != nil {
		utils.Errorf("Service error in UpdateBuildingsInfo: %v", err)
		// 如果 Service 返回错误，直接返回 500 和错误信息
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("更新建筑物时发生错误: %v", err),
		})
		return
	}

	// 如果 Service 成功，result 就是期望的 map[string]interface{} (例如 {"success": true, ...})
	c.JSON(http.StatusOK, result)
}
