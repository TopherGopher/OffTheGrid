package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"rsc.io/pdf"
)

func main() {
	_, err := readPdf("workbook.pdf") // Read local pdf file
	if err != nil {
		panic(err)
	}
	// fmt.Println(content)
	return
}

func readPdf(path string) (string, error) {
	r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("could not open file: %+v", err)
	}
	fullText := ""
	// Loop over the pages/text
	for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {
		for _, text := range r.Page(pageNum).Content().Text {
			fullText += text.S
		}
	}
	serviceHunks := splitIntoHunks(fullText)
	for _, hunk := range serviceHunks {
		parseServiceHunk(hunk)
	}

	return fmt.Sprintf("%+v", ""), nil
}

func splitIntoHunks(fullText string) []string {
	// Get the index of the first Service in the string
	firstServiceIndex := strings.Index(fullText, "Service:")
	// Skip the preface to the first service
	// preface := text[0:firstServiceIndex]
	serviceHunks := []string{}
	i := firstServiceIndex
	for i < len(fullText)-1 {
		// +1 to serviceIndex so we don't catch this 'Service:' with the find
		// Then we +1 since the string was shorter by a character
		// Then we +i because it's a relative pointer
		nextServiceIndex := strings.Index(fullText[i+1:], "Service:") + 1 + i
		logrus.Infof("%s\n", fullText[i:nextServiceIndex])
		serviceHunks = append(serviceHunks, fullText[i:nextServiceIndex])
		if i == nextServiceIndex {
			break
		}
		i = nextServiceIndex
	}

	return serviceHunks
}

// Example hunk:
// Service:ZabasearchUpdated:October2018Website:zabasearch.comRemovalLink:NonePrivacyPolicy:zabasearch.com/privacy.phpEmailAddress:info@zabasearch.com,response@zabasearch.comRequirements:FaxsubmissionNotes:Sendyourcustomopt-outrequestformviafaxto425-974-6194.Date:___________Response:_______________________VerifiedRemoval:______
func parseServiceHunk(hunk string) WorkbookService {
	re := regexp.MustCompile("Service:(?P<service>.*)(?msi:Updated:(?P<updated>.*))+(?:Website:(?P<website>[[:graph:]]+))+(?:RemovalLink:(?P<removalLink>[[:graph:]]+))+(?:PrivacyPolicy:(?P<privacyPolicy>[[:graph:]]+))+(?:EmailAddress:(?P<email>.*))+(?:Requirements:(?P<requirements>[[:graph:]]+))+Faxsubmission(?:Notes:(?P<notes>[[:graph:]]+))*(?P<other>.*)")
	groupNames := re.SubexpNames()
	for matchNum, match := range re.FindAllStringSubmatch(hunk, -1) {
		for groupIdx, group := range match {
			name := groupNames[groupIdx]
			if name == "" {
				name = "*"
			}
			fmt.Printf("#%d text: '%s', group: '%s'\n", matchNum, group, name)
		}
	}

	// fmt.Printf("%+v\n", hostGroupMatch)
	return WorkbookService{
		ServiceName: "",
	}
}

type WorkbookService struct {
	ServiceName   string
	Requirements  string
	Email         string
	PrivacyPolicy string
	Updated       string
	Notes         string
}
