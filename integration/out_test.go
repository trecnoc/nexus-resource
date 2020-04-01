package integration_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	neturl "net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/trecnoc/nexus-resource/models"
	"github.com/trecnoc/nexus-resource/out"
)

var _ = Describe("out", func() {
	var (
		command   *exec.Cmd
		stdin     *bytes.Buffer
		session   *gexec.Session
		sourceDir string

		expectedExitStatus int
	)

	BeforeEach(func() {
		var err error
		sourceDir, err = ioutil.TempDir("", "nexus_out_integration_test")
		Ω(err).ShouldNot(HaveOccurred())

		stdin = &bytes.Buffer{}
		expectedExitStatus = 0

		command = exec.Command(outPath, sourceDir)
		command.Stdin = stdin
	})

	AfterEach(func() {
		err := os.RemoveAll(sourceDir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		<-session.Exited
		Expect(session.ExitCode()).To(Equal(expectedExitStatus))
	})

	Context("with a file to upload", func() {
		var groupPrefix string

		BeforeEach(func() {
			groupPrefix = "/out-request-files"

			err := ioutil.WriteFile(filepath.Join(sourceDir, "glob-file-to-upload"), []byte("contents"), 0755)
			Ω(err).ShouldNot(HaveOccurred())

			outRequest := out.Request{
				Source: models.Source{
					URL:        url,
					Repository: repository,
					Username:   username,
					Password:   password,
					Group:      groupPrefix,
				},
				Params: out.Params{
					File: "glob-*",
				},
			}

			err = json.NewEncoder(stdin).Encode(&outRequest)
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "glob-file-to-upload"))
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("uploads the file to the correct bucket and outputs the version", func() {
			// Let Nexus time to index for search
			time.Sleep(2 * time.Second)

			nexuspaths, err := nexusclient.ListFiles(repository, groupPrefix)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(nexuspaths).Should(ConsistOf(filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "glob-file-to-upload")))

			reader := bytes.NewBuffer(session.Buffer().Contents())

			var response out.Response
			err = json.NewDecoder(reader).Decode(&response)
			Ω(err).ShouldNot(HaveOccurred())

			u, _ := neturl.Parse(url)
			u.Path = path.Join(u.Path, "repository", repository, "out-request-files/glob-file-to-upload")

			Ω(response).Should(Equal(out.Response{
				Version: models.Version{
					Path: filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "glob-file-to-upload"),
				},
				Metadata: []models.MetadataPair{
					{
						Name:  "filename",
						Value: "glob-file-to-upload",
					},
					{
						Name:  "url",
						Value: u.String(),
					},
				},
			}))
		})
	})
})
