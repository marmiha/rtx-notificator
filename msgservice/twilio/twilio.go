package twilio

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type SmsClient struct{
	SID string
	AuthToken string
	NumFrom string
	NumsTo []string
	httpClient *http.Client
}

func NewSmsClient(SID string, authToken string, numFrom string, numTo ...string) *SmsClient {
	client := SmsClient{
		SID: SID,
		AuthToken: authToken,
		NumFrom: numFrom,
		NumsTo: numTo,
		httpClient: &http.Client{},
	}
	return &client
}

func (c SmsClient) Send(msg string) ([]string, []error) {
	var errs []error
	var result []string

	for _, num := range c.NumsTo {
		urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + c.SID + "/Messages.json"

		v := url.Values{}
		v.Set("To", num)
		v.Set("From", c.NumFrom)
		v.Set("Body", msg)

		rb := strings.NewReader(v.Encode())
		req, err := http.NewRequest("POST", urlStr, rb)

		if err != nil {
			fmt.Printf("Could not create new http request for number %v\n", num)
			continue
		}

		req.SetBasicAuth(c.SID, c.AuthToken)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errs = append(errs, errors.New(fmt.Sprintf("%v err -> %v", num, err)))
			continue
		}

		if resp.StatusCode == http.StatusCreated {
			result = append(result, fmt.Sprintf("\"%v\"  ->  %v", num, resp.Status))
		} else {
			errs = append(errs, errors.New(fmt.Sprintf("\"%v\"  ->  %v", num, resp.Status)))
		}
	}

	return result, errs
}

