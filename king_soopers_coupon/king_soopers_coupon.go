package couponpusher

import (
	"fmt"
	"strings"

	offthegrid "github.com/TopherGopher/OffTheGrid"
	"github.com/chromedp/cdproto/cdp"
	"github.com/sirupsen/logrus"
)

type CouponBase struct {
	LoginURL             string
	CouponURL            string
	couponButtonSelector string
	AccountCouponURL     string
	webDriver            *offthegrid.WebDriver
	formAnalyzer         *offthegrid.Analyzer
	log                  *logrus.Logger
}

type CouponInterface interface {
	Login() error
}

type KingSoopersCoupon struct {
	CouponBase
}

func NewKingSoopersCoupon() *KingSoopersCoupon {
	wd := offthegrid.NewWebDriver()
	headless := false
	wd.Init(headless)
	return &KingSoopersCoupon{
		CouponBase: CouponBase{
			LoginURL:             "https://www.kingsoopers.com/signin?redirectUrl=/cl/coupons/",
			CouponURL:            "https://www.kingsoopers.com/cl/coupons",
			AccountCouponURL:     "https://www.kingsoopers.com/cl/mycoupons/",
			couponButtonSelector: "div.CouponCard-buttonContainer.CouponCard-row.CouponCard-button > button",
			webDriver:            wd,
			formAnalyzer:         offthegrid.NewAnalyzer(),
			log:                  logrus.New(),
		},
	}
}

// Teardown tears down the web driver
func (cb *CouponBase) Teardown() {
	cb.webDriver.Teardown()
}

// Login to King Soopers site - session persists in webDriver
//*[@id="content"]/section/div/section[4]/div/div[2]/div/div/div/div[2]/div/div/div/ul/li[1]/div/div/div[4]/div[2]/button
func (cb *CouponBase) Login() (err error) {
	return cb.webDriver.Login(
		cb.LoginURL,
		"//*[@id=\"SignIn-emailInput\"]", "topher@develops.guru",
		"//*[@id=\"SignIn-passwordInput\"]", "MyPASSWORD",
		"//*[@id=\"SignIn-submitButton\"]",
		"//*[@id=\"content\"]/section/div/section[4]/div/div[2]/div/div/div/div[2]/div/div/div/ul/li[1]/div/div/div[4]/div[2]/button",
		cb.couponButtonSelector,
	)
}

// CouponsAreAvailable returns True if there are coupon buttons available for clicking
func (cb *CouponBase) CouponsAreAvailable() bool {
	var couponBody string
	var elems []*cdp.Node
	if elems = cb.webDriver.FetchElements(cb.couponButtonSelector); len(elems) == 0 {
		cb.log.Debug("No coupon buttons were found")
		return false
	}

	// if len(cb.webDriver.BlacklistCoupons) >= len(elems) {
	// 	foundSomethingClickable := false
	// 	for _, elem := range elems {
	// 		// TODO: Add this check so we don't loop forever
	// 		// we need to compute the text here
	// 		if _, ok := cb.webDriver.BlacklistCoupons[]; !ok {
	// 			foundSomethingClickable = true
	// 			break
	// 		}
	// 	}
	// 	if !foundSomethingClickable {
	// 		cb.log.Debug("No non-blacklist elements were found on the page. Not an error.")
	// 		return false
	// 	}
	// }

	if strings.Contains(elems[0].AttributeValue("class"), "CouponButton-maxLimitReached") {
		cb.log.Warn("No more coupons can be added to this account - King Soopers server-side Limit")
		return false
	}
	cb.log.WithField("couponBody", couponBody).Debug("Found a coupon button")
	return true
}

// RemoveAllCoupons removes all coupons from your account
func (cb *CouponBase) RemoveAllCoupons() (err error) {
	if err = cb.webDriver.GoToPage(cb.AccountCouponURL); err != nil {
		return err
	}
	if !cb.CouponsAreAvailable() {
		cb.log.Info("No remove coupon buttons could be found to click.")
		return fmt.Errorf("no remove coupon buttons available")
	}
	confirm := false
	for cb.CouponsAreAvailable() {
		// While coupons are available, click everything on screen
		// then refresh
		if err = cb.webDriver.ClickAllButtons(cb.couponButtonSelector, confirm); err != nil {
			cb.log.WithField("error", err).Error("There was an issue clicking the buttons")
			return err
		}
		if err = cb.webDriver.ReloadPage(); err != nil {
			cb.log.WithField("error", err).Error("There was an issue reloading the page")
			return err
		}
	}
	return nil
}

// DoIt calls Login and clicks any relevant coupon buttons
func (cb *CouponBase) DoIt() (err error) {
	if err = cb.Login(); err != nil {
		cb.log.WithField("error", err).Error("There was an issue logging in")
		return err
	}
	confirm := true
	for cb.CouponsAreAvailable() {
		// While coupons are available, click everything on screen
		// then refresh
		if err = cb.webDriver.ClickAllButtons(cb.couponButtonSelector, confirm); err != nil {
			cb.log.WithField("error", err).Error("There was an issue clicking the buttons")
			return err
		}
		break
		// TODO: Bring back reload code when CouponsAreAvailable references
		// proper map key
		// if err = cb.webDriver.ReloadPage(); err != nil {
		// 	cb.log.WithField("error", err).Error("There was an issue reloading the page")
		// 	return err
		// }
		// cb.log.Debug("The coupon page has been reloaded")
	}
	cb.log.Info("No more coupon buttons were found to click.")
	return nil
}

// Login to CVS and put every offer on card
