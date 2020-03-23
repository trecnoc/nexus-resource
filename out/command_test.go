package out_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/trecnoc/nexus-resource/out"

	"github.com/onsi/gomega/gbytes"
	"github.com/trecnoc/nexus-resource/fakes"
	"github.com/trecnoc/nexus-resource/models"
)

var _ = Describe("Out Command", func() {
	Describe("running the command", func() {
		var (
			tmpPath   string
			sourceDir string
			request   Request

			stderr      *gbytes.Buffer
			nexusclient *fakes.FakeNexusClient
			command     *Command
		)

		BeforeEach(func() {
			var err error
			tmpPath, err = ioutil.TempDir("", "out_command")
			Ω(err).ShouldNot(HaveOccurred())

			sourceDir = filepath.Join(tmpPath, "source")
			err = os.MkdirAll(sourceDir, 0755)
			Ω(err).ShouldNot(HaveOccurred())

			request = Request{
				Source: models.Source{
					URL:        "http://nexus-url.com",
					Repository: "repository-name",
					Username:   "user",
					Password:   "password",
				},
			}

			nexusclient = &fakes.FakeNexusClient{}
			nexusclient.URLStub = func(repositoryName string, remotePath string) string {
				return "http://nexus-url.com/" + filepath.Join(repositoryName, remotePath)
			}
			stderr = gbytes.NewBuffer()
			command = NewCommand(stderr, nexusclient)
		})

		AfterEach(func() {
			stderr.Close()
			err := os.RemoveAll(tmpPath)
			Ω(err).ShouldNot(HaveOccurred())
		})

		createFile := func(path string) {
			fullPath := filepath.Join(sourceDir, path)
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			Ω(err).ShouldNot(HaveOccurred())

			file, err := os.Create(fullPath)
			Ω(err).ShouldNot(HaveOccurred())
			file.Close()
		}

		Describe("finding files to upload with File param", func() {
			It("does not error if there is a single match", func() {
				request.Source.Group = "/files"
				request.Params.File = "a/*.tgz"
				createFile("a/file.tgz")

				_, err := command.Run(sourceDir, request)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("errors if there are no matches", func() {
				request.Source.Group = "/files"
				request.Params.File = "b/*.tgz"
				createFile("a/file1.tgz")
				createFile("a/file2.tgz")

				_, err := command.Run(sourceDir, request)
				Ω(err).Should(HaveOccurred())
			})

			It("errors if there are more than one match", func() {
				request.Source.Group = "/files"
				request.Params.File = "a/*.tgz"
				createFile("a/file1.tgz")
				createFile("a/file2.tgz")

				_, err := command.Run(sourceDir, request)
				Ω(err).Should(HaveOccurred())
			})
		})

		Describe("uploading files", func() {
			It("uploads to the root group", func() {
				request.Source.Group = "/"
				request.Params.File = "my/*.tgz"
				createFile("my/special-file.tgz")

				response, err := command.Run(sourceDir, request)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(nexusclient.UploadFileCallCount()).Should(Equal(1))
				repositoryName, group, remoteFileName, localPath := nexusclient.UploadFileArgsForCall(0)
				Ω(repositoryName).Should(Equal("repository-name"))
				Ω(group).Should(Equal("/"))
				Ω(remoteFileName).Should(Equal("special-file.tgz"))
				Ω(localPath).Should(Equal(filepath.Join(sourceDir, "my/special-file.tgz")))

				Ω(nexusclient.URLCallCount()).Should(Equal(1))
				repositoryName, remotePath := nexusclient.URLArgsForCall(0)
				Ω(repositoryName).Should(Equal("repository-name"))
				Ω(remotePath).Should(Equal("special-file.tgz"))

				Ω(response.Version.Path).Should(Equal("special-file.tgz"))

				Ω(response.Metadata[0].Name).Should(Equal("filename"))
				Ω(response.Metadata[0].Value).Should(Equal("special-file.tgz"))

				Ω(response.Metadata[1].Name).Should(Equal("url"))
				Ω(response.Metadata[1].Value).Should(Equal("http://nexus-url.com/repository-name/special-file.tgz"))
			})

			It("uploads to a non-root group", func() {
				request.Source.Group = "/files"
				request.Params.File = "a/*.tgz"
				createFile("a/file.tgz")

				response, err := command.Run(sourceDir, request)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(nexusclient.UploadFileCallCount()).Should(Equal(1))
				repositoryName, directory, remoteFileName, localPath := nexusclient.UploadFileArgsForCall(0)
				Ω(repositoryName).Should(Equal("repository-name"))
				Ω(directory).Should(Equal("/files"))
				Ω(remoteFileName).Should(Equal("file.tgz"))
				Ω(localPath).Should(Equal(filepath.Join(sourceDir, "a/file.tgz")))

				Ω(nexusclient.URLCallCount()).Should(Equal(1))
				repositoryName, remotePath := nexusclient.URLArgsForCall(0)
				Ω(repositoryName).Should(Equal("repository-name"))
				Ω(remotePath).Should(Equal("files/file.tgz"))

				Ω(response.Version.Path).Should(Equal("files/file.tgz"))

				Ω(response.Metadata[0].Name).Should(Equal("filename"))
				Ω(response.Metadata[0].Value).Should(Equal("file.tgz"))

				Ω(response.Metadata[1].Name).Should(Equal("url"))
				Ω(response.Metadata[1].Value).Should(Equal("http://nexus-url.com/repository-name/files/file.tgz"))
			})
		})

	})
})
