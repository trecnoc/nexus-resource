package out

import "github.com/trecnoc/nexus-resource/models"

type Request struct {
	Source models.Source `json:"source"`
	Params Params        `json:"params"`
}

type Params struct {
	File string `json:"file"`
}

type Response struct {
	Version  models.Version        `json:"version"`
	Metadata []models.MetadataPair `json:"metadata"`
}
