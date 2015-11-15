package application_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/application"
)

var _ = Describe("Filesystem Config Store", func() {
	var tempDir string
	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "tubes-unit-test-")
		Expect(err).NotTo(HaveOccurred())
	})

	It("stores data on the filesystem", func() {
		By("initializing a store at one location")
		rootDir1 := tempDir
		instance1 := application.FilesystemConfigStore{RootDir: rootDir1}

		By("writing some data using that store")
		Expect(instance1.Set("some/path/to/a/thing", []byte("some data"))).To(Succeed())

		By("moving the data somewhere else on the filesystem")
		rootDir2, err := ioutil.TempDir("", "tubes-unit-test-dest-")
		Expect(err).NotTo(HaveOccurred())
		destDir := filepath.Join(rootDir2, "new-location")
		Expect(os.Rename(rootDir1, destDir)).To(Succeed())

		By("creating another store at the new location")
		instance2 := application.FilesystemConfigStore{RootDir: destDir}

		By("reading the data out of that store")
		value, err := instance2.Get("some/path/to/a/thing")
		Expect(err).NotTo(HaveOccurred())

		By("checking that the data is intact")
		Expect(value).To(Equal([]byte("some data")))
	})

	Context("when the store location is a relative path", func() {
		It("succeeds", func() {
			os.Chdir(tempDir)
			store := application.FilesystemConfigStore{RootDir: "."}

			someData := []byte(fmt.Sprintf("%x", rand.Int63()))
			Expect(store.Set("some/key", someData)).To(Succeed())

			Expect(store.Get("some/key")).To(Equal(someData))
		})
	})

	Context("when the root dir does not exist", func() {
		It("should return an error", func() {
			store := application.FilesystemConfigStore{RootDir: filepath.Join(tempDir, "does-not-exist/")}

			Expect(store.Set("key", []byte("value"))).To(BeAssignableToTypeOf(&os.PathError{}))
			_, err := store.Get("key")
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})
	})

	Context("when the root dir is a file", func() {
		It("should return an error", func() {
			badPath := filepath.Join(tempDir, "foo")
			Expect(ioutil.WriteFile(badPath, []byte("whatever"), 0666)).To(Succeed())
			store := application.FilesystemConfigStore{RootDir: badPath}

			Expect(store.Set("key", []byte("value"))).To(MatchError(ContainSubstring("to be a directory")))
		})
	})

	Context("when file does not exist", func() {
		It("should return an error", func() {
			store := application.FilesystemConfigStore{RootDir: tempDir}

			_, err := store.Get("key")
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})
	})

	Context("when the key contains an invalid character", func() {
		It("should return an error", func() {
			store := application.FilesystemConfigStore{RootDir: tempDir}

			invalidKey := "some/path" + string([]byte{0x00}) + "more/path"
			err := store.Set(invalidKey, []byte("some data"))
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))

			invalidKey = "some/path" + string([]byte{0x00})
			err = store.Set(invalidKey, []byte("some data"))
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})
	})
})
