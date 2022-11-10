package offthegrid

import "fmt"

// FormFields is both a hold
type FormFields struct {
	FirstName            string
	LastName             string
	HouseNumber          int
	StreetName           string
	ApartmentNumber      int
	Email                string
	PhoneNumber          string
	CellNumber           string
	SocialSecurityNumber int
	DriversLicenseNumber int
	SubmitButtonXPath    string
}

// Person is a helper for holding metadata about an individual.
// Combinations of the fields are used to fill fields
type Person = FormFields

func (p *Person) FullName() string {
	return fmt.Sprintf("%s %s", p.FirstName, p.LastName)
}
