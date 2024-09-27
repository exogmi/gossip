package models

import "github.com/google/uuid"

// generateUniqueID generates a unique identifier using UUID
func generateUniqueID() string {
	return uuid.New().String()
}
