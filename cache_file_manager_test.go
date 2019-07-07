package offthegrid

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func setupCacheTest() {
	logrus.SetLevel(logrus.DebugLevel)
}

func cleanupTestCache() {
	os.RemoveAll(scrapeCacheFolder)
}

func TestCreateCacheFile(t *testing.T) {
	setupCacheTest()
	defer cleanupTestCache()
	assert := assert.New(t)
	cfm := NewCacheFileManager()
	assert.NotNil(cfm)
	err := cfm.CachePageLocally("Some Fake Content", "https://some.fake.url")
	assert.NoError(err)
	assert.FileExists(GetCachedFilePath("https://some.fake.url"))
}

func TestRetrieveCacheFile(t *testing.T) {
	defer cleanupTestCache()
	assert := assert.New(t)
	cfm := NewCacheFileManager()
	assert.NotNil(cfm)
	// Check to make sure non-existence is handled gracefully
	content, err := cfm.FetchLocalCachedPage("https://does.not.exist")
	assert.NoError(err)
	assert.Equal(0, len(content))
	err = cfm.CachePageLocally("Some Fake Content", "https://some.fake.url")
	assert.NoError(err)
	content, err = cfm.FetchLocalCachedPage("https://some.fake.url")
	assert.NoError(err)
	assert.Equal("Some Fake Content", content)
}

func TestCacheFileExists(t *testing.T) {
	defer cleanupTestCache()
	assert := assert.New(t)
	cfm := NewCacheFileManager()
	assert.NotNil(cfm)
	err := cfm.CachePageLocally("Some Fake Content", "https://some.fake.url")
	assert.NoError(err)

	assert.True(cfm.CacheFileExists("https://some.fake.url"))
	assert.False(cfm.CacheFileExists("https://does.not.exist"))
}

func TestPageHasChanged(t *testing.T) {
	defer cleanupTestCache()
	assert := assert.New(t)
	cfm := NewCacheFileManager()
	assert.NotNil(cfm)
	err := cfm.CachePageLocally("Some Fake Content", "https://some.fake.url")
	assert.NoError(err)

	assert.Equal("", cfm.GetCachedFileSha("https://does.not.exist"))

	assert.Equal("da320156a63e7608eefe3ddf5f512f04", cfm.GetCachedFileSha("https://some.fake.url"))

	newSha := cfm.GetShaFromString("Some Fake Content")
	assert.NoError(err)
	assert.Equal("da320156a63e7608eefe3ddf5f512f04", newSha)

	assert.False(cfm.PageHasChanged("Some Fake Content", "https://some.fake.url"))
	assert.True(cfm.PageHasChanged("This content shouldn't match", "https://some.fake.url"))
}
