package pkg

import (
	"fmt"
	"net/http"
	"os"
)

func (networkErr *StatusError) networkCheckBlaze() {
	//fmt.Println("Executing networkCheckBlaze...")
	blazeNetworkCheck := []string{"https://a.blazemeter.com", "https://data.blazemeter.com", "https://mock.blazemeter.com", "https://auth.blazemeter.com", "https://storage.blazemeter.com", "https://bard.blazemeter.com", "https://tdm.blazemeter.com", "https://analytics.blazemeter.com"}
	client := &http.Client{}
	for _, url := range blazeNetworkCheck {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client req error for: %v\n", url)
			networkErr.BlazeNetworkStatus = append(networkErr.BlazeNetworkStatus, map[string]error{statement: err})
			fmt.Println(statement, err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client.Do error for: %v\n", url)
			networkErr.BlazeNetworkStatus = append(networkErr.BlazeNetworkStatus, map[string]error{statement: err})
			fmt.Println(statement, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err := fmt.Errorf("network error for %s, status code: %d", url, resp.StatusCode)
			networkErr.BlazeNetworkStatus = append(networkErr.BlazeNetworkStatus, map[string]error{url: err})
			fmt.Println("Network error:", err)
			continue
		}
		fmt.Printf("Network check passed for: %s, with status %v\n", url, resp.StatusCode)
	}
}

func (networkErr *StatusError) networkCheckImageRegistry() {

	imageRegistryCheck := os.Getenv("DOCKER_REGISTRY")
	fmt.Println(imageRegistryCheck)

	client := &http.Client{}
	req, err := http.NewRequest("GET", imageRegistryCheck, nil)
	if err != nil {
		//statement := fmt.Sprintf("This is a Go HTTP client req error for: %v", err)
		networkErr.ImageRegistryNetworkStatus = err
	}
	resp, err := client.Do(req)
	if err != nil {
		//statement := fmt.Sprintf("This is a Go HTTP client.Do error for: %v", imageRegistryCheck)
		networkErr.ImageRegistryNetworkStatus = err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err := fmt.Errorf("network error connecting to %s, status code: %d", imageRegistryCheck, resp.StatusCode)
		networkErr.ImageRegistryNetworkStatus = err
	}
	fmt.Printf("Network check passed for: %s, with status %v\n", imageRegistryCheck, resp.StatusCode)

}

func (networkErr *StatusError) networkCheckThirdParty() {
	//	networkErr := &StatusError{}
	//	var imageRegistryCheck string
	//	imageRegistryCheck = os.Getenv("DOCKER_REGISTRY")
	//	if imageRegistryCheck != "gcr.io/verdant-bulwark-278" {
	//
	//	}
	thirdPartyNetworkCheck := []string{"https://pypi.org/", "https://storage.googleapis.com", "https://hub.docker.com", "https://index.docker.io"}
	client := &http.Client{}

	for _, url := range thirdPartyNetworkCheck {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client req error for: %v\n", url)
			networkErr.ThirdPartyNetworkStatus = append(networkErr.ThirdPartyNetworkStatus, map[string]error{statement: err})
			break
		}
		resp, err := client.Do(req)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client.Do error for: %v\n", url)
			networkErr.ThirdPartyNetworkStatus = append(networkErr.ThirdPartyNetworkStatus, map[string]error{statement: err})
			break
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err := fmt.Errorf("network error, status code: %d", resp.StatusCode)
			networkErr.ThirdPartyNetworkStatus = append(networkErr.ThirdPartyNetworkStatus, map[string]error{url: err})
			continue
		}
		fmt.Printf("Network check passed for: %s, with status %v\n", url, resp.StatusCode)
	}
}
