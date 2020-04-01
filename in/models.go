package in

import "github.com/trecnoc/nexus-resource/models"

// Request struct for the In command
type Request struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
	Params  Params         `json:"params"`
}

// Params struct for the Out command
type Params struct {
	Unpack       bool `json:"unpack"`
	SkipDownload bool `json:"skip_download"`
}

// Response struct of the In command
type Response struct {
	Version  models.Version        `json:"version"`
	Metadata []models.MetadataPair `json:"metadata"`
}
