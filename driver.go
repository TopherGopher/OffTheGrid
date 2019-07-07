package offthegrid

// import "github.com/tebeka/selenium"
import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/chromedp/chromedp"
)

// WebDriver is a helper for hanging
type WebDriver struct {
	chromeDpContext context.Context
	cacheManager    *CacheFileManager
	log             *logrus.Logger
}

// NewWebDriver creates the skeleton for a new web driver.
// It should almost always be followed by Init() unless testing
func NewWebDriver() *WebDriver {
	return &WebDriver{
		log:          logrus.New(),
		cacheManager: NewCacheFileManager(),
	}
}

// Init initializes the WebDriver and populates the
// skeleton.
func (wd *WebDriver) Init() (err error) {
	// f, err := os.OpenFile("driver.log", os.O_CREATE|os.O_RDWR, 0666)
	// if err != nil {
	// 	fmt.Printf("error opening file: %v", err)
	// }
	// wd.log.SetOutput(f)
	wd.log.SetLevel(logrus.DebugLevel)
	err = wd.CreateChromeDPDriver()
	if err != nil {
		wd.log.WithField("error", err).Error("Could not create a new ChromeDP driver")
		return err
	}
	wd.cacheManager = NewCacheFileManager()
	return nil
}

// Teardown stops all related tasks related to the driver
func (wd *WebDriver) Teardown() {
	err := chromedp.Cancel(wd.chromeDpContext)
	if err != nil {
		wd.log.Panicf("Could not cancel everything: %+v", err)
	}
}

// CreateChromeDPDriver spawns a new window
func (wd *WebDriver) CreateChromeDPDriver() error {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-hang-monitor", true),
		// chromedp.UserAgent()
	} // append( //chromedp.DefaultExecAllocatorOptions[:],

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer cancel()

	// also set up a custom logger
	taskCtx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	// defer cancel()
	wd.chromeDpContext = taskCtx

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		return err
	}
	return nil
}

// GetInnerHTMLOfElement returns the raw HTML of a web element using
// the chromedp driver. This is useful for forms with javascript
// rendering.
func (wd *WebDriver) GetInnerHTMLOfElement(url string, elementName string) (body string, err error) {
	err = chromedp.Run(wd.chromeDpContext,
		chromedp.Navigate(url),
		// chromedp.Text(`#pkg-overview`, &body, chromedp.NodeVisible, chromedp.ByID),
		chromedp.InnerHTML(elementName, &body), //, chromedp.NodeVisible),
	)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not fetch page")
	}
	return body, err
}

// GetFullPageHTML fetches the entire HTML for a given URL using the
// http.Get() request.
// If the page is a fairly straight-up form, this should work fine.
func (wd *WebDriver) GetFullPageHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not retrieve the HTML of the web page")
		return "", err
	}
	if resp.StatusCode > 300 {
		wd.log.WithFields(logrus.Fields{
			"statusCode": resp.StatusCode,
			"body":       resp.Body,
		}).Error("Could not retrieve the HTML of the web page")
		return "", fmt.Errorf("bad status code")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not read the response payload")
		return "", err
	}
	return string(body), nil
}

// SiteHasChangedSinceLastPull returns True if the site content
// differs from the local content. A True value is returned
// if there was an issue fetching the page as it's possible the site
// went away.
func (wd *WebDriver) SiteHasChangedSinceLastPull(url string, saveIfNew bool) bool {
	body, err := wd.GetFullPageHTML(url)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not fetch the requested page's HTML")
		return true
	}
	// Compare the body we just got to what's on disk
	hasChanged := wd.cacheManager.PageHasChanged(body, url)
	if hasChanged && saveIfNew {
		// If we've been told to save if there are changes, then add it to cache
		err = wd.cacheManager.CachePageLocally(body, url)
		if err != nil {
			wd.log.WithField("error", err).Error("Could not cache the page locally")
			return true
		}
	}

	return hasChanged
}

// GetAndCacheSite fetches a site and saves it to disk without
// performing any collision checks. Force overwrite.
func (wd *WebDriver) GetAndCacheSite(url string) (err error) {
	pageHTML, err := wd.GetFullPageHTML(url)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not fetch the requested page's HTML")
		return err
	}
	err = wd.cacheManager.CachePageLocally(pageHTML, url)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not fetch the requested page's HTML")
		return err
	}
	return nil
}
