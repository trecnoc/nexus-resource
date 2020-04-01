package out

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/trecnoc/nexus-resource"
	"github.com/trecnoc/nexus-resource/models"
)

// Command struct for Out
type Command struct {
	nexusclient nexusresource.NexusClient
	stderr      io.Writer
}

// NewCommand creates a new Out command
func NewCommand(stderr io.Writer, nexusclient nexusresource.NexusClient) *Command {
	return &Command{
		stderr:      stderr,
		nexusclient: nexusclient,
	}
}

// Run the command
func (command *Command) Run(sourceDir string, request Request) (Response, error) {
	if ok, message := request.Source.IsValid(); !ok {
		return Response{}, errors.New(message)
	}

	localPath, err := command.match(request.Params, sourceDir)
	if err != nil {
		return Response{}, err
	}
	repositoryName := request.Source.Repository
	group := request.Source.Group
	localFileName := filepath.Base(localPath)

	err = command.nexusclient.UploadFile(
		repositoryName,
		group,
		localFileName,
		localPath,
	)
	if err != nil {
		return Response{}, err
	}

	var remotePath string
	if request.Source.Group == "/" {
		remotePath = localFileName
	} else {
		remotePath = strings.TrimPrefix(request.Source.Group, "/") + "/" + localFileName
	}

	version := models.Version{}
	version.Path = remotePath

	return Response{
		Version:  version,
		Metadata: command.metadata(repositoryName, localFileName, remotePath),
	}, nil
}

func (command *Command) match(params Params, sourceDir string) (string, error) {
	var matches []string
	var err error
	var pattern string

	pattern = params.File
	matches, err = filepath.Glob(filepath.Join(sourceDir, pattern))

	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found for pattern: %s", pattern)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("more than one match found for pattern: %s\n%v", pattern, matches)
	}

	return matches[0], nil
}

func (command *Command) metadata(repositoryName string, localFileName string, remotePath string) []models.MetadataPair {
	metadata := []models.MetadataPair{
		models.MetadataPair{
			Name:  "filename",
			Value: localFileName,
		},
	}

	metadata = append(metadata, models.MetadataPair{
		Name:  "url",
		Value: command.nexusclient.URL(repositoryName, remotePath),
	})

	return metadata
}
