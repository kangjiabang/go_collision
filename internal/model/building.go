// internal/model/building.go
package model

// Building represents a building record from the database.
// Using pointers for fields that can be NULL in the database.
type Building struct {
	// BuildingID is typically not NULL, so keep it as a value type.
	// If it can be NULL, use *int64.
	BuildingID int64 `json:"building_id" db:"building_id"`

	// BuildingName can be NULL, so use *string.
	BuildingName *string `json:"building_name,omitempty" db:"building_name"` // omitempty prevents sending null/undefined in JSON if nil

	// Geom is usually text, but if ST_AsText can return NULL (rare, but possible with invalid geometries?), use *string.
	Geom *string `json:"geom,omitempty" db:"geom"`

	// BuildingHeight could potentially be NULL, use *float64.
	BuildingHeight *float64 `json:"building_height,omitempty" db:"building_height"`
}
