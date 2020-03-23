package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/trecnoc/nexus-resource/in"
	"github.com/trecnoc/nexus-resource/models"
)

var _ = Describe("in", func() {
	var (
		command            *exec.Cmd
		inRequest          in.Request
		stdin              *bytes.Buffer
		session            *gexec.Session
		destDir            string
		expectedExitStatus int
	)

	BeforeEach(func() {
		var err error
		destDir, err = ioutil.TempDir("", "nexus_in_integration_test")
		Ω(err).ShouldNot(HaveOccurred())

		stdin = &bytes.Buffer{}
		expectedExitStatus = 0

		command = exec.Command(inPath, destDir)
		command.Stdin = stdin
	})

	AfterEach(func() {
		err := os.RemoveAll(destDir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		var err error

		err = json.NewEncoder(stdin).Encode(inRequest)
		Ω(err).ShouldNot(HaveOccurred())

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		<-session.Exited
		Expect(session.ExitCode()).To(Equal(expectedExitStatus))
	})

	Context("when the given version contains a path", func() {
		var groupPrefix string

		BeforeEach(func() {
			groupPrefix = "in-request-files"
			inRequest = in.Request{
				Source: models.Source{
					URL:        url,
					Repository: repository,
					Username:   username,
					Password:   password,
					Regexp:     filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "some-file-(.*)"),
				},
				Version: models.Version{
					Path: filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "some-file-2"),
				},
			}

			tempFile, err := ioutil.TempFile("", "file-to-upload")
			Ω(err).ShouldNot(HaveOccurred())
			tempFile.Close()

			for i := 1; i <= 3; i++ {
				err = ioutil.WriteFile(tempFile.Name(), []byte(fmt.Sprintf("some-file-%d", i)), 0755)
				Ω(err).ShouldNot(HaveOccurred())

				err = nexusclient.UploadFile(repository, groupPrefix, fmt.Sprintf("some-file-%d", i), tempFile.Name())
				Ω(err).ShouldNot(HaveOccurred())
			}

			err = os.Remove(tempFile.Name())
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			for i := 1; i <= 3; i++ {
				err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(groupPrefix, "/"), fmt.Sprintf("some-file-%d", i)))
				Ω(err).ShouldNot(HaveOccurred())
			}
		})

		It("downloads the file", func() {
			reader := bytes.NewBuffer(session.Out.Contents())

			var response in.Response
			err := json.NewDecoder(reader).Decode(&response)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(response).Should(Equal(in.Response{
				Version: models.Version{
					Path: "in-request-files/some-file-2",
				},
				Metadata: []models.MetadataPair{
					{
						Name:  "filename",
						Value: "some-file-2",
					},
					{
						Name:  "url",
						Value: url + "/repository/" + repository + "/in-request-files/some-file-2",
					},
					{
						Name:  "sha",
						Value: "39a09cfff0422cae1cc3a8ddc40581ad51adc61f",
					},
				},
			}))

			Ω(filepath.Join(destDir, "some-file-2")).Should(BeARegularFile())
			contents, err := ioutil.ReadFile(filepath.Join(destDir, "some-file-2"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(contents).Should(Equal([]byte("some-file-2")))

			Ω(filepath.Join(destDir, "version")).Should(BeARegularFile())
			versionContents, err := ioutil.ReadFile(filepath.Join(destDir, "version"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(versionContents).Should(Equal([]byte("2")))

			Ω(filepath.Join(destDir, "url")).Should(BeARegularFile())
			urlContents, err := ioutil.ReadFile(filepath.Join(destDir, "url"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(urlContents).Should(Equal([]byte(url + "/repository/" + repository + "/in-request-files/some-file-2")))

			Ω(filepath.Join(destDir, "sha")).Should(BeARegularFile())
			shaContents, err := ioutil.ReadFile(filepath.Join(destDir, "sha"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(shaContents).Should(Equal([]byte("39a09cfff0422cae1cc3a8ddc40581ad51adc61f")))
		})

	})
})
