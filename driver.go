package offthegrid

// import "github.com/tebeka/selenium"
import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/andybalholm/cascadia"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/eiannone/keyboard"
	"golang.org/x/net/html"
)

// WebDriver is a helper for hanging
type WebDriver struct {
	chromeDpContext  context.Context
	cacheManager     *CacheFileManager
	log              *logrus.Logger
	BlacklistCoupons map[string]bool
}

// NewWebDriver creates the skeleton for a new web driver.
// It should almost always be followed by Init() unless testing
func NewWebDriver() *WebDriver {
	return &WebDriver{
		log:              logrus.New(),
		cacheManager:     NewCacheFileManager(),
		BlacklistCoupons: map[string]bool{},
	}
}

// Init initializes the WebDriver and populates the
// skeleton.
func (wd *WebDriver) Init(headless bool) (err error) {
	// f, err := os.OpenFile("driver.log", os.O_CREATE|os.O_RDWR, 0666)
	// if err != nil {
	// 	fmt.Printf("error opening file: %v", err)
	// }
	// wd.log.SetOutput(f)
	wd.log.SetLevel(logrus.DebugLevel)
	err = wd.CreateChromeDPDriver(headless)
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
		// unlinkat /var/folders/jr/l48h_6kj71n4y2q4gqjrld2w0000gn/T/chromedp-runner260946925/Default: directory not empty
		if strings.HasPrefix(err.Error(), "unlinkat") {
			// There was an issue cleaning up a tmp folder. Whack it manually.
			errComponents := strings.SplitAfter(err.Error(), " ")
			path := errComponents[1]
			path = path[0 : len(path)-1]
			if err2 := os.RemoveAll(path); err2 != nil {
				wd.log.Panicf("Failed to perform fallback cleanup: %v", err2)
			}
		} else {
			wd.log.Panicf("Could not cleanly teardown chromedp: %+v", err)
		}
	}
}

// CreateChromeDPDriver spawns a new window
func (wd *WebDriver) CreateChromeDPDriver(headless bool) error {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-hang-monitor", true),
		// chromedp.UserAgent()
	} // append( //chromedp.DefaultExecAllocatorOptions[:],
	if headless {
		opts = append(opts, chromedp.Headless)
	}

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

// GoToPage navigates to the given URL
func (wd *WebDriver) GoToPage(url string) error {
	return chromedp.Run(wd.chromeDpContext, chromedp.Navigate(url))
}

// GetInnerHTMLOfElement returns the raw HTML of a web element using
// the chromedp driver. This is useful for forms with javascript
// rendering and to maintain session.
func (wd *WebDriver) GetInnerHTMLOfElement(elementName string) (body string, err error) {
	err = chromedp.Run(wd.chromeDpContext,
		chromedp.InnerHTML(elementName, &body), //, chromedp.NodeVisible),
	)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not fetch page")
	}
	return body, err
}

// FetchElements returns a slice of all nodes that match the selector
// nil is returned in the case that nothing is found.
func (wd *WebDriver) FetchElements(selector string) (elements []*cdp.Node) {
	if err := chromedp.Run(wd.chromeDpContext, chromedp.Nodes(selector, &elements)); err != nil {
		wd.log.WithField("error", err).Debug("Could not find any coupon buttons")
		return nil
	}
	return elements
}

// FetchElement returns the first node that matches the selector
// nil is returned in the case that nothing is found.
func (wd *WebDriver) FetchElement(selector string) (element *cdp.Node) {
	elements := wd.FetchElements(selector)
	if elements == nil || len(elements) == 0 {
		return nil
	}
	return elements[0]
}

// Login logs the user into a site and waits for the element with the 'waitForThis' selector to become
// visible before returning. If waitForThis is an empty string, no waiting is performed.
func (wd *WebDriver) Login(loginURL, userFieldName, username, passwordFieldName, password, submitButton, waitForThis, couponButtons string) (err error) {
	err = chromedp.Run(wd.chromeDpContext,
		chromedp.Navigate(loginURL),
		chromedp.SendKeys(userFieldName, username),
		chromedp.SendKeys(passwordFieldName, password),
		chromedp.Click(submitButton), // Submit the login page
	)
	if err != nil {
		wd.log.WithField("error", err).Error("Could not fetch page")
		return err
	}
	if waitForThis != "" {
		// Wait until this element is visible to finish logging on
		if err = chromedp.Run(wd.chromeDpContext, chromedp.WaitVisible(waitForThis)); err != nil {
			wd.log.WithField("error", err).Error("Test coupon button never became visible")
			return err
		}
		// Wait for the footer (looking for an arbitrary link to ensure page is built)
		if err = chromedp.Run(wd.chromeDpContext, chromedp.WaitVisible("footer > nav > section:nth-child(5) > a:nth-child(6)")); err != nil {
			wd.log.WithField("error", err).Error("Test coupon button never became visible")
			return err
		}
	}

	wd.log.Debug("Login has finished processing")

	return nil
}

// ReloadPage reloads the current webpage
func (wd *WebDriver) ReloadPage() (err error) {
	if err = chromedp.Run(wd.chromeDpContext, chromedp.Reload()); err != nil {
		return err
	}
	return nil
}

// ClickButton clicks a button given a selector
func (wd *WebDriver) ClickButton(buttonSelector string) (err error) {
	couponButtonNode := wd.FetchElement(buttonSelector)
	if couponButtonNode == nil {
		err = fmt.Errorf("could not find a button to click")
		return err
	}
	if err = chromedp.Run(wd.chromeDpContext, chromedp.Click(buttonSelector)); err != nil {
		return err
	}
	return nil
}

// ClickAllButtons clicks all buttons that match the buttonSelector
// Used by couponpusher to click LoadToCard
func (wd *WebDriver) ClickAllButtons(buttonSelector string, confirm bool) (err error) {
	var couponText string
	// Sleep for a sec while the buttons populate
	wd.log.Debug("Scrolling 'infinitely'")
	chromedp.Run(
		wd.chromeDpContext,
		chromedp.Sleep(time.Second*5),
		// Scroll to the footer to account for infinite scroll
		chromedp.ScrollIntoView("footer > nav > section:nth-child(5) > a:nth-child(6)"),
		chromedp.Sleep(time.Second*5),
		chromedp.ScrollIntoView("footer > nav > section:nth-child(5) > a:nth-child(6)"),
		chromedp.Sleep(time.Second*5),
		chromedp.ScrollIntoView("footer > nav > section:nth-child(5) > a:nth-child(6)"),
		chromedp.Sleep(time.Second*5),
		chromedp.ScrollIntoView("footer > nav > section:nth-child(5) > a:nth-child(6)"),
		chromedp.Sleep(time.Second*5),
		chromedp.ScrollIntoView("footer > nav > section:nth-child(5) > a:nth-child(6)"),
		chromedp.Sleep(time.Second*5),
	)
	wd.log.Debug("Done scrolling 'infinitely'")

	// Get all the "couponCard" divs
	couponDivSelector := "div.Card > div.CouponCard"
	couponDivNodes := wd.FetchElements(couponDivSelector)
	couponButtonNodes := wd.FetchElements(buttonSelector)
	if len(couponDivNodes) == 0 {
		err = fmt.Errorf("no coupon divs were found")
		return err
	}
	if len(couponDivNodes) != len(couponButtonNodes) {
		err = fmt.Errorf("a different number of descriptor nodes were found than coupon button nodes")
		return err
	}

	wd.log.WithFields(logrus.Fields{
		"numBlackList":  len(wd.BlacklistCoupons),
		"numCouponDivs": len(couponDivNodes),
		"blacklist":     wd.BlacklistCoupons,
		"couponDivs":    couponDivNodes,
	}).Debug("We should have something available to load")

	// Activate a keyboard scanner to read from STDIN
	if err = keyboard.Open(); err != nil {
		wd.log.Debug("Could not open keyboard")
		return err
	}

	defer keyboard.Close()
	for i, couponDivNode := range couponDivNodes {
		xpath := couponDivNode.FullXPathByID()

		label := couponButtonNodes[i].AttributeValue("aria-label")
		// This will be something like "Baby,"
		dataCategory := couponDivNode.AttributeValue("data-category")
		if strings.Contains(dataCategory, "Baby") {
			// Skip anything in the baby category
			wd.BlacklistCoupons[couponText] = true
			continue
		}
		couponDivNode, err := wd.GetInnerHTMLOfElement(couponDivNode.FullXPath())
		if err != nil {
			return err
		}

		doc, err := html.Parse(strings.NewReader(couponDivNode))
		if err != nil {
			return err
		}

		couponTextNode := cascadia.MustCompile(".CouponCard-img").MatchFirst(doc)
		if couponTextNode == nil {
			err = fmt.Errorf("coupon text node not found")
			return err
		}

		for _, attr := range couponTextNode.Attr {
			//Image Save $1.00 on 2 Angie's BOOMCHICKAPOP®\u200b Ready to Eat Popcorn, Click on this image to view more info in coupon modal
			if attr.Key == "aria-label" {
				couponText = attr.Val
				couponText = strings.TrimPrefix(couponText, "Image ")
				couponText = strings.TrimSuffix(couponText, ", Click on this image to view more info in coupon modal")
				break
			}
		}
		if _, ok := wd.BlacklistCoupons[couponText]; ok {
			// If this is part of the blacklist, continue
			wd.log.WithField("text", couponText).Debug("Found existing blacklist entry")
			continue
		}
		if confirm {
			if !strings.Contains(label, "Load to Card") {
				// If this isn't a coupon, or the coupon has already been loaded
				// then we don't want to click the button
				continue
			}
			fmt.Printf("------------------------\nCategory: %s\nText: %s\n------------------------\n", dataCategory, couponText)
			fmt.Print("Would you like to load this coupon? (y/n): ")

			answer, key, err := keyboard.GetKey()
			if err != nil {
				panic(err)
			}
			if key == keyboard.KeyCtrlC || key == keyboard.KeyCtrlD || key == keyboard.KeyCtrlV {
				panic("CTRL+C or CTRL+D or CTRL+V was pressed")
			}
			fmt.Printf("%s\n", string(answer))
			if answer != 'y' {
				fmt.Print("\tSkipping and blacklisting\n")
				wd.BlacklistCoupons[couponText] = true
				continue
			}

			// Add a newline
			fmt.Println("")
		}

		wd.log.WithFields(logrus.Fields{
			"xpath": xpath,
			"label": label,
		}).Debug("I found a button to click")
		if err = chromedp.Run(wd.chromeDpContext, chromedp.Click(couponButtonNodes[i].FullXPath())); err != nil {
			wd.log.WithFields(logrus.Fields{
				"error": err,
				"xpath": xpath,
				"label": label,
			}).Error("Could not click the coupon button")
			return err
		}
	}
	return nil
}

// GetFullPageHTML fetches the entire HTML for a given URL using the
// http.Get() request. It does not consume the session/cookies from WebDriver/chromedp.
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
