// Manages local file caching
package offthegrid

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	url_package "net/url"
	"os"

	"github.com/sirupsen/logrus"
)

type CacheFileManager struct {
	log *logrus.Logger
	// TODO: Cache disk checksums in memory?
	// cacheLock *sync.RWMutex
	// diskShaMap map[string]string
}

func NewCacheFileManager() *CacheFileManager {
	// Try to make the cache directory - ignore any errors
	os.Mkdir(scrapeCacheFolder, 0755)
	return &CacheFileManager{
		log: logrus.New(),
	}
}

var scrapeCacheFolder = "scraped_pages"

// GetCachedFilePath converts a URL to a relative file path
func GetCachedFilePath(url string) string {
	parsedURL, err := url_package.Parse(url)
	if err != nil {
		// If we weren't able to parse the URL, then just
		// use the URL as it is
		return fmt.Sprintf("%s/%s", scrapeCacheFolder, url)
	}
	// Otherwise use the hostname as the file path
	return fmt.Sprintf("%s/%s", scrapeCacheFolder, parsedURL.Host)
}

// CachePageLocally caches a page's content into a local file
func (cfm *CacheFileManager) CachePageLocally(pageHTML, url string) error {
	err := ioutil.WriteFile(GetCachedFilePath(url), []byte(pageHTML), 0644)
	return err
}

// FetchLocalCachedPage returns the HTML content of a local page.
// If no page could be found, an empty string is returned
func (cfm *CacheFileManager) FetchLocalCachedPage(url string) (string, error) {
	fileBytes, err := ioutil.ReadFile(GetCachedFilePath(url))
	if err != nil && os.IsNotExist(err) {
		// If the file doesn't exist, then return an emapty string
		// This is expected if we haven't cached anything before
		return "", nil
	}
	if err != nil {
		cfm.log.WithField("error", err).Error("Could not fetch local page from cache for an unexpected reason")
		return "", err
	}
	return string(fileBytes), nil
}

// CacheFileExists determines if a local version of a cached file
// is present.
func (cfm *CacheFileManager) CacheFileExists(url string) bool {
	if _, err := os.Stat(GetCachedFilePath(url)); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetShaFromString calculates the MD5 checksum of a string. This is useful for
// calculating SHAs in memory.
func (cfm *CacheFileManager) GetShaFromString(fileContents string) string {
	shaHasher := md5.New()
	shaHasher.Write([]byte(fileContents))
	return hex.EncodeToString(shaHasher.Sum(nil))
}

// GetCachedFileSha calculates the MD5 checksum of a file's contents
// and returns the SHA. In the case that no file exists, an empty string
// without an error is returned.
func (cfm *CacheFileManager) GetCachedFileSha(url string) string {
	var f *os.File
	// Open the cache file
	f, err := os.Open(GetCachedFilePath(url))
	if err != nil && os.IsNotExist(err) {
		// If the file doesn't exist, just return an empty string
		return ""
	}
	if err != nil {
		cfm.log.WithField("error", err).Error("Unable to open file to calculate the md5 checksum")
		return ""
	}
	defer f.Close()

	shaHasher := md5.New()
	if numBytesCopied, err := io.Copy(shaHasher, f); err != nil {
		cfm.log.WithField("error", err).Error("Unable to copy the file pointer to the md5 hasher")
		return ""
	} else if numBytesCopied == 0 {
		cfm.log.Warn("No content was found in the cached file")
		return ""
	}
	return hex.EncodeToString(shaHasher.Sum(nil))
}

// PageHasChanged compares in-memory content to a cached file's contents
func (cfm *CacheFileManager) PageHasChanged(fileContents, url string) bool {
	return cfm.GetCachedFileSha(url) != cfm.GetShaFromString(fileContents)
}
