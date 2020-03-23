package check

import (
	"github.com/trecnoc/nexus-resource/models"
)

type Request struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
}

type Response []models.Version
