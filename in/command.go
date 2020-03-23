package in

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/trecnoc/nexus-resource"
	"github.com/trecnoc/nexus-resource/models"
	"github.com/trecnoc/nexus-resource/versions"
)

var ErrMissingPath = errors.New("missing path in request")

type MetadataProvider struct {
	nexusClient nexusresource.NexusClient
}

func (provider *MetadataProvider) GetURL(request Request, remotePath string) string {
	return provider.nexusURL(request, remotePath)
}

func (provider *MetadataProvider) nexusURL(request Request, remotePath string) string {
	return provider.nexusClient.URL(request.Source.Repository, remotePath)
}

func (provider *MetadataProvider) GetSHA(request Request, remotePath string) string {
	return provider.nexusSHA(request, remotePath)
}

func (provider *MetadataProvider) nexusSHA(request Request, remotePath string) string {
	return provider.nexusClient.SHA(request.Source.Repository, remotePath)
}

type Command struct {
	nexusclient      nexusresource.NexusClient
	metadataProvider MetadataProvider
}

func NewCommand(nexusclient nexusresource.NexusClient) *Command {
	return &Command{
		nexusclient: nexusclient,
		metadataProvider: MetadataProvider{
			nexusClient: nexusclient,
		},
	}
}

func (command *Command) Run(destinationDir string, request Request) (Response, error) {
	if ok, message := request.Source.IsValid(); !ok {
		return Response{}, errors.New(message)
	}

	err := os.MkdirAll(destinationDir, 0755)
	if err != nil {
		return Response{}, err
	}

	var remotePath string
	var versionNumber string
	var url string
	var sha string

	if request.Version.Path == "" {
		return Response{}, ErrMissingPath
	}

	remotePath = request.Version.Path
	extraction, ok := versions.Extract(remotePath, request.Source.Regexp)
	if !ok {
		return Response{}, fmt.Errorf("regex does not match provided version: %#v", request.Version)
	}

	versionNumber = extraction.VersionNumber

	if !request.Params.SkipDownload {
		err = command.downloadFile(
			request.Source.Repository,
			remotePath,
			destinationDir,
			path.Base(remotePath),
		)
		if err != nil {
			return Response{}, err
		}

		if request.Params.Unpack {
			destinationPath := filepath.Join(destinationDir, path.Base(remotePath))
			mime := archiveMimetype(destinationPath)
			if mime == "" {
				return Response{}, fmt.Errorf("not an archive: %s", destinationPath)
			}

			err = extractArchive(mime, destinationPath)
			if err != nil {
				return Response{}, err
			}
		}
	}

	url = command.metadataProvider.GetURL(request, remotePath)
	err = command.writeURLFile(destinationDir, url)
	if err != nil {
		return Response{}, err
	}

	sha = command.metadataProvider.GetSHA(request, remotePath)
	err = command.writeSHAFile(destinationDir, sha)
	if err != nil {
		return Response{}, err
	}

	err = command.writeVersionFile(versionNumber, destinationDir)
	if err != nil {
		return Response{}, err
	}

	metadata := command.metadata(remotePath, url, sha)

	return Response{
		Version: models.Version{
			Path: remotePath,
		},
		Metadata: metadata,
	}, nil
}

func (command *Command) writeURLFile(destDir string, url string) error {
	return ioutil.WriteFile(filepath.Join(destDir, "url"), []byte(url), 0644)
}

func (command *Command) writeSHAFile(destDir string, sha string) error {
	return ioutil.WriteFile(filepath.Join(destDir, "sha"), []byte(sha), 0644)
}

func (command *Command) writeVersionFile(versionNumber string, destDir string) error {
	return ioutil.WriteFile(filepath.Join(destDir, "version"), []byte(versionNumber), 0644)
}

func (command *Command) downloadFile(repositoryName string, remotePath string, destinationDir string, destinationFile string) error {
	localPath := filepath.Join(destinationDir, destinationFile)

	return command.nexusclient.DownloadFile(
		repositoryName,
		remotePath,
		localPath,
	)
}

func (command *Command) metadata(remotePath string, url string, sha string) []models.MetadataPair {
	remoteFilename := filepath.Base(remotePath)

	metadata := []models.MetadataPair{
		models.MetadataPair{
			Name:  "filename",
			Value: remoteFilename,
		},
	}

	if url != "" {
		metadata = append(metadata, models.MetadataPair{
			Name:  "url",
			Value: url,
		})
	}

	if sha != "" {
		metadata = append(metadata, models.MetadataPair{
			Name:  "sha",
			Value: sha,
		})
	}

	return metadata
}

func extractArchive(mime, filename string) error {
	destDir := filepath.Dir(filename)

	err := inflate(mime, filename, destDir)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %s", err)
	}

	if mime == "application/gzip" || mime == "application/x-gzip" {
		fileInfos, err := ioutil.ReadDir(destDir)
		if err != nil {
			return fmt.Errorf("failed to read dir: %s", err)
		}

		if len(fileInfos) != 1 {
			return fmt.Errorf("%d files found after gunzip; expected 1", len(fileInfos))
		}

		filename = filepath.Join(destDir, fileInfos[0].Name())
		mime = archiveMimetype(filename)
		if mime == "application/x-tar" {
			err = inflate(mime, filename, destDir)
			if err != nil {
				return fmt.Errorf("failed to extract archive: %s", err)
			}
		}
	}

	return nil
}
