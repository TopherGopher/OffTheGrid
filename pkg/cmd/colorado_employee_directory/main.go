package main

webAddress := "https://apps.colorado.gov/apps/oit/directory/searchResults.jsf"


type EmployeeScraperBase struct {
	// LoginURL             string
	// CouponURL            string
	// couponButtonSelector string
	// AccountCouponURL     string
	webDriver            *offthegrid.WebDriver
	formAnalyzer         *offthegrid.Analyzer
	log                  *logrus.Logger
}

func NewKingSoopersCoupon() *KingSoopersCoupon {
	wd := offthegrid.NewWebDriver()
	headless := false
	wd.Init(headless)
	return &EmployeeScraperBase{
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

func main() {

}