package out

import "github.com/trecnoc/nexus-resource/models"

// Request struct for the Out command
type Request struct {
	Source models.Source `json:"source"`
	Params Params        `json:"params"`
}

// Params struct for the Out command
type Params struct {
	File string `json:"file"`
}

// Response struct of the Out command
type Response struct {
	Version  models.Version        `json:"version"`
	Metadata []models.MetadataPair `json:"metadata"`
}
