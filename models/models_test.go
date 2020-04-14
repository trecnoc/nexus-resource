package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/trecnoc/nexus-resource/models"
)

var _ = Describe("Models", func() {
	Context("when calling isValid()", func() {
		Context("when the source is valid", func() {
			It("validates a valid source", func() {
				var source = models.Source{
					URL:        "http://nexus-url.com",
					Repository: "repository-name",
					Username:   "user",
					Password:   "password",
					Group:      "/group",
					Regexp:     "group/a(.*).tgz",
					Debug:      false,
				}

				ok, err := source.IsValid()
				Ω(ok).Should(BeTrue())
				Ω(err).Should(Equal(""))
			})

			It("defaults debug to false", func() {
				var source = models.Source{
					URL:        "http://nexus-url.com",
					Repository: "repository-name",
					Username:   "user",
					Password:   "password",
					Group:      "/group",
					Regexp:     "group/a(.*).tgz",
				}

				Ω(source.Debug).Should(BeFalse())
			})
		})

		Context("when the source is invalid", func() {
			It("validates missing URL", func() {

				var source = models.Source{
					Repository: "repository-name",
					Username:   "user",
					Password:   "password",
					Group:      "/group",
					Regexp:     "group/a(.*).tgz",
				}

				ok, err := source.IsValid()
				Ω(ok).Should(BeFalse())
				Ω(err).Should(Equal("url must be specified"))
			})

			It("validates missing Repository", func() {
				var source = models.Source{
					URL:      "http://nexus-url.com",
					Username: "user",
					Password: "password",
					Group:    "/group",
					Regexp:   "group/a(.*).tgz",
				}

				ok, err := source.IsValid()
				Ω(ok).Should(BeFalse())
				Ω(err).Should(Equal("repository must be specified"))
			})

			It("validates missing username", func() {
				var source = models.Source{
					URL:        "https://nexus-url.com",
					Repository: "repository-name",
					Password:   "password",
					Group:      "/group",
					Regexp:     "group/a(.*).tgz",
				}

				ok, err := source.IsValid()
				Ω(ok).Should(BeFalse())
				Ω(err).Should(Equal("username must be specified"))
			})

			It("validates missing password", func() {
				var source = models.Source{
					URL:        "https://nexus-url.com",
					Repository: "repository-name",
					Username:   "admin",
					Group:      "/group",
					Regexp:     "group/a(.*).tgz",
				}

				ok, err := source.IsValid()
				Ω(ok).Should(BeFalse())
				Ω(err).Should(Equal("password must be specified"))
			})

			It("validates invalid Directory", func() {
				var source = models.Source{
					URL:        "http://nexus-url.com",
					Repository: "repository-name",
					Username:   "user",
					Password:   "password",
					Group:      "group",
					Regexp:     "group/a(.*).tgz",
				}

				ok, err := source.IsValid()
				Ω(ok).Should(BeFalse())
				Ω(err).Should(Equal("group must start with '/'"))
			})

			It("validates invalid Regex", func() {
				var source = models.Source{
					URL:        "http://nexus-url.com",
					Repository: "repository-name",
					Group:      "/group",
					Username:   "user",
					Password:   "password",
					Regexp:     "/group/a(.*).tgz",
				}

				ok, err := source.IsValid()
				Ω(ok).Should(BeFalse())
				Ω(err).Should(Equal("regexp should not start with '/'"))
			})
		})
	})
})
