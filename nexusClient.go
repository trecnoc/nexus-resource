package nexusresource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/trecnoc/nexus-resource/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/FakeNexusClient.go --fake-name FakeNexusClient . NexusClient

// NexusClient Interface
type NexusClient interface {
	ListFiles(repositoryName string, group string) ([]string, error)
	DownloadFile(repositoryName string, name string, localPath string) error
	UploadFile(repositoryName string, group string, remoteFilename string, localPath string) error
	DeleteFile(repositoryName string, name string) error
	URL(repositoryName string, name string) string
	SHA(repositoryName string, name string) string
}

type nexusclient struct {
	httpClient *http.Client
	nexusURL   string
	username   string
	password   string
}

// NewNexusClient creates and returns an NexusClient
func NewNexusClient(nexusURL string, username string, password string) NexusClient {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &nexusclient{
		httpClient: httpClient,
		nexusURL:   nexusURL,
		username:   username,
		password:   password,
	}
}

func (client *nexusclient) ListFiles(repositoryName string, group string) ([]string, error) {
	entries, err := client.getRepositoryGroupContent(repositoryName, group)

	if err != nil {
		return []string{}, err
	}

	paths := make([]string, 0, len(entries))

	for _, entry := range entries {
		paths = append(paths, entry.Name)
	}
	return paths, nil
}

func (client *nexusclient) DownloadFile(repositoryName string, name string, localPath string) error {
	var url string

	url = client.URL(repositoryName, name)

	resp, err := client.doGetRequest(url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	localFile, err := os.Create(localPath + ".tmp")
	if err != nil {
		return err
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, resp.Body)
	if err != nil {
		return err
	}

	err = os.Rename(localPath+".tmp", localPath)
	if err != nil {
		return err
	}

	return nil
}

func (client *nexusclient) UploadFile(repositoryName string, group string, remoteFilename string, localPath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("raw.asset1", filepath.Base(localPath))
	if err != nil {
		return err
	}
	_, _ = io.Copy(part, localFile)
	_ = writer.WriteField("raw.directory", group)
	_ = writer.WriteField("raw.asset1.filename", remoteFilename)

	err = writer.Close()
	if err != nil {
		return err
	}

	u, _ := url.Parse(client.nexusURL)
	u.Path = path.Join(u.Path, "service/rest/v1/components")
	q, _ := url.ParseQuery(u.RawQuery)
	q.Add("repository", repositoryName)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("UploadFile: received invalid status code %d", resp.StatusCode)
	}

	return nil
}

func (client *nexusclient) DeleteFile(repositoryName string, name string) error {
	item, err := client.getRepositoryItem(repositoryName, name)
	if err != nil {
		return err
	}

	u, _ := url.Parse(client.nexusURL)
	u.Path = path.Join(u.Path, "service/rest/v1/components", item.ID)

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}

	if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("DeleteFile: received invalid status code %d", resp.StatusCode)
	}
	resp.Body.Close()
	return nil
}

func (client *nexusclient) URL(repositoryName string, name string) string {
	u, _ := url.Parse(client.nexusURL)
	u.Path = path.Join(u.Path, "repository", repositoryName, name)
	return u.String()
}

func (client *nexusclient) SHA(repositoryName string, name string) string {
	var sha string

	item, err := client.getRepositoryItem(repositoryName, name)
	if err == nil {
		sha = item.Assets[0].Checksum.Sha1
	}

	return sha
}

func (client *nexusclient) doGetRequest(requestURL string, parameters map[string]string) (*http.Response, error) {
	u, _ := url.Parse(requestURL)
	if parameters != nil || len(parameters) > 0 {
		q, _ := url.ParseQuery(u.RawQuery)
		for key, value := range parameters {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return nil, fmt.Errorf("doGetResquest: non-successful status code received %d", resp.StatusCode)
	}
	return resp, nil
}

func (client *nexusclient) doGetRequestPath(requestPath string, parameters map[string]string) (*http.Response, error) {
	u, _ := url.Parse(client.nexusURL)
	u.Path = path.Join(u.Path, requestPath)

	return client.doGetRequest(u.String(), parameters)
}

func (client *nexusclient) getRepositoryGroupContent(repositoryName string, group string) (map[string]models.RepositoryItem, error) {
	repositoryItems := map[string]models.RepositoryItem{}
	continuation := ""

	getContents := func() (repositoryItems models.RespositoryItems, err error) {
		var parameters map[string]string

		parameters = make(map[string]string)
		parameters["repository"] = repositoryName
		parameters["group"] = group

		if continuation != "" {
			parameters["continuationToken"] = continuation
		}

		response, err := client.doGetRequestPath("service/rest/v1/search", parameters)
		if err != nil {
			return repositoryItems, err
		}
		defer response.Body.Close()

		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&repositoryItems)

		return
	}

	items := make([]models.RepositoryItem, 0)
	for {
		resp, err := getContents()
		if err != nil {
			return repositoryItems, err
		}

		items = append(items, resp.Items...)

		if resp.ContinuationToken == "" {
			break
		}

		continuation = resp.ContinuationToken
	}

	for _, item := range items {
		repositoryItems[item.Name] = item
	}

	return repositoryItems, nil
}

func (client *nexusclient) getRepositoryItem(repositoryName string, name string) (models.RepositoryItem, error) {
	var item models.RepositoryItem
	var parameters map[string]string

	parameters = make(map[string]string)
	parameters["repository"] = repositoryName
	parameters["name"] = name

	response, err := client.doGetRequestPath("/service/rest/v1/search", parameters)
	if err != nil {
		return item, err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	var items models.RespositoryItems
	err = decoder.Decode(&items)
	if err != nil {
		return item, err
	}

	if len(items.Items) != 1 {
		return item, fmt.Errorf("getRepositoryItem: expected 1 Component got %d", len(items.Items))
	} else if len(items.Items[0].Assets) != 1 {
		return item, fmt.Errorf("getRepositoryItem: Component should only have 1 Asset, contains %d", len(items.Items[0].Assets))
	}

	item = items.Items[0]

	return item, nil
}
