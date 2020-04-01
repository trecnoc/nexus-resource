package check

import (
	"github.com/trecnoc/nexus-resource/models"
)

// Request struct for the Check command
type Request struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
}

// Response struct of the Check command
type Response []models.Version
