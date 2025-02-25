package main

import (
	"fmt"
	"net/http"
)

func netWorkCheckBlaze() {
	networkErr := &StatusError{}
	networkCheck := []string{"a.blazemeter.com, data.blazemeter.com, mock.blazemeter.com, auth.blazemeter.com, storage.blazemeter.com, bard.blazemeter.com, tdm.blazemeter.com, analytics.blazemeter.com"}
	client := &http.Client{}
	for _, url := range networkCheck {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client req error for: %v", url)
			networkErr.NetworkStatus = append(networkErr.NetworkStatus, map[string]error{statement: err})
			break
		}
		resp, err := client.Do(req)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client.Do error for: %v", url)
			networkErr.NetworkStatus = append(networkErr.NetworkStatus, map[string]error{statement: err})
			break
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err := fmt.Errorf("network error, status code: %d", resp.StatusCode)
			networkErr.NetworkStatus = append(networkErr.NetworkStatus, map[string]error{url: err})
			continue
		}
	}

}
