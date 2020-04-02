package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nexuxclient", func() {
	var (
		tempDir     string
		tempFile    *os.File
		runtime     string
		groupPrefix string
	)
	BeforeEach(func() {
		var err error
		groupPrefix = "/nexusclient-tests"

		runtime = fmt.Sprintf("%d", time.Now().Unix())

		tempDir, err = ioutil.TempDir("", "nexus-upload-dir")
		Ω(err).ShouldNot(HaveOccurred())

		tempFile, err = ioutil.TempFile(tempDir, "file-to-upload")
		Ω(err).ShouldNot(HaveOccurred())

		tempFile.Write([]byte("hello-" + runtime))
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Ω(err).ShouldNot(HaveOccurred())

		paths, err := nexusclient.ListFiles(repository, groupPrefix)

		Ω(err).ShouldNot(HaveOccurred())
		for _, path := range paths {
			err := nexusclient.DeleteFile(repository, path)
			Ω(err).ShouldNot(HaveOccurred())
		}
	})

	It("can interact with a repository", func() {
		err := nexusclient.UploadFile(repository, groupPrefix, "file-to-upload-1", tempFile.Name())
		Ω(err).ShouldNot(HaveOccurred())

		err = nexusclient.UploadFile(repository, groupPrefix, "file-to-upload-2", tempFile.Name())
		Ω(err).ShouldNot(HaveOccurred())

		err = nexusclient.UploadFile(repository, groupPrefix, "file-to-upload-2", tempFile.Name())
		Ω(err).ShouldNot(HaveOccurred())

		err = nexusclient.UploadFile(repository, groupPrefix, "file-to-upload-3", tempFile.Name())
		Ω(err).ShouldNot(HaveOccurred())

		// Let Nexus time to index for search
		time.Sleep(1 * time.Second)

		paths, err := nexusclient.ListFiles(repository, groupPrefix)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(paths).Should(ConsistOf([]string{
			filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "file-to-upload-1"),
			filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "file-to-upload-2"),
			filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "file-to-upload-3"),
		}))

		err = nexusclient.DownloadFile(repository, filepath.Join(strings.TrimPrefix(groupPrefix, "/"), "file-to-upload-1"), filepath.Join(tempDir, "downloaded-file"))
		Ω(err).ShouldNot(HaveOccurred())

		read, err := ioutil.ReadFile(filepath.Join(tempDir, "downloaded-file"))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(read).Should(Equal([]byte("hello-" + runtime)))
	})

})
