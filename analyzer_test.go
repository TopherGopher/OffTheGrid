package offthegrid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFormFields(t *testing.T) {
	assert := assert.New(t)
	driver := NewWebDriver()
	htmlBody, err := driver.GetFullPageHTML("https://google.com")
	assert.NoError(err)
	analyzer := NewAnalyzer()
	analyzer.FindFormFields(htmlBody)
	assert.True(false)
}
