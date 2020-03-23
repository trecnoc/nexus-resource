package in

import "github.com/trecnoc/nexus-resource/models"

type Request struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
	Params  Params         `json:"params"`
}

type Params struct {
	Unpack       bool `json:"unpack"`
	SkipDownload bool `json:"skip_download"`
}

type Response struct {
	Version  models.Version        `json:"version"`
	Metadata []models.MetadataPair `json:"metadata"`
}
