package check

import (
	"errors"

	"github.com/trecnoc/nexus-resource"
	"github.com/trecnoc/nexus-resource/models"
	"github.com/trecnoc/nexus-resource/versions"
)

type Command struct {
	nexusclient nexusresource.NexusClient
}

func NewCommand(nexusclient nexusresource.NexusClient) *Command {
	return &Command{
		nexusclient: nexusclient,
	}
}

func (command *Command) Run(request Request) (Response, error) {
	if ok, message := request.Source.IsValid(); !ok {
		return Response{}, errors.New(message)
	}

	extractions := versions.GetRepositoryItemVersions(command.nexusclient, request.Source)

	if len(extractions) == 0 {
		return nil, nil
	}

	lastVersion, matched := versions.Extract(request.Version.Path, request.Source.Regexp)
	if !matched {
		return latestVersion(extractions), nil
	} else {
		return newVersions(lastVersion, extractions), nil
	}
}

func latestVersion(extractions versions.Extractions) Response {
	lastExtraction := extractions[len(extractions)-1]
	return []models.Version{{Path: lastExtraction.Path}}
}

func newVersions(lastVersion versions.Extraction, extractions versions.Extractions) Response {
	response := Response{}

	for _, extraction := range extractions {
		if extraction.Version.Compare(lastVersion.Version) >= 0 {
			version := models.Version{
				Path: extraction.Path,
			}
			response = append(response, version)
		}
	}

	return response
}
