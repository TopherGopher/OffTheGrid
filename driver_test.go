package offthegrid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWebDriver(t *testing.T) {
	assert := assert.New(t)
	wd := NewWebDriver()
	defer wd.Teardown()
	err := wd.Init()
	assert.NoError(err)
}

func TestGetPageWithChromeDP(t *testing.T) {
	assert := assert.New(t)
	wd := NewWebDriver()
	defer wd.Teardown()
	err := wd.Init()
	assert.NoError(err)
	body, err := wd.GetPageBody("https://google.com")
	assert.NoError(err)
	assert.GreaterOrEqual(len(body), 1000)
}

func TestGetPageWithHTTP(t *testing.T) {
	assert := assert.New(t)
	wd := NewWebDriver()
	defer wd.Teardown()
	err := wd.Init()
	assert.NoError(err)
	body, err := wd.GetFullPageHTML("https://google.com")
	assert.NoError(err)
	assert.GreaterOrEqual(len(body), 1000)
}

func TestSiteHasChangedSinceLastPull(t *testing.T) {
	assert := assert.New(t)
	wd := NewWebDriver()
	defer wd.Teardown()
	defer cleanupTestCache()
	err := wd.Init()
	assert.NoError(err)
	// The site isn't cached yet, so we should see that the site has
	// changed.
	// The true we are passing in says to save it to local cache
	assert.True(wd.SiteHasChangedSinceLastPull("https://example.com", true))
	// Which means if we check again, we'll see disk matches remote
	assert.False(wd.SiteHasChangedSinceLastPull("https://example.com", true))
}
