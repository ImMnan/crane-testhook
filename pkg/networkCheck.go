package pkg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func (statusError *StatusError) networkCheckBlaze() {
	//fmt.Println("Executing networkCheckBlaze...")
	blazeNetworkCheck := []string{"https://a.blazemeter.com", "https://data.blazemeter.com", "https://mock.blazemeter.com", "https://auth.blazemeter.com", "https://storage.blazemeter.com", "https://bard.blazemeter.com", "https://tdm.blazemeter.com", "https://analytics.blazemeter.com"}
	client := &http.Client{}
	for _, url := range blazeNetworkCheck {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client req error for: %v\n", url)
			statusError.BlazeNetworkStatus = append(statusError.BlazeNetworkStatus, map[string]error{statement: err})
			fmt.Printf("\n[%s] %v", time.Now().Format("2006-01-02 15:04:05"), statement)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client.Do error for: %v\n", url)
			statusError.BlazeNetworkStatus = append(statusError.BlazeNetworkStatus, map[string]error{statement: err})
			fmt.Printf("\n[%s] %v", time.Now().Format("2006-01-02 15:04:05"), statement)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err := fmt.Errorf("network error for %s, status code: %d", url, resp.StatusCode)
			statusError.BlazeNetworkStatus = append(statusError.BlazeNetworkStatus, map[string]error{url: err})
			fmt.Printf("\n[%s] %v", time.Now().Format("2006-01-02 15:04:05"), err)
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("\nError reading response body:", err)
			} else {
				fmt.Println("\nResponse body:", string(body))
			}
			continue
		}

		fmt.Printf("\n[%s],Network check passed for: %s, with status %v\n", time.Now().Format("2006-01-02 15:04:05"), url, resp.StatusCode)
	}
}

func (statusError *StatusError) networkCheckImageRegistry() {
	imageRegistryCheck := os.Getenv("DOCKER_REGISTRY")
	imageRegistry := fmt.Sprintf("https://%s", imageRegistryCheck)

	client := &http.Client{}
	req, err := http.NewRequest("GET", imageRegistry, nil)
	if err != nil {
		statusError.ImageRegistryNetworkStatus = err
		fmt.Printf("\n[%s] This is a Go HTTP client req error for: %s, %v", time.Now().Format("2006-01-02 15:04:05"), imageRegistryCheck, err)
		return // Exit the function if request creation fails
	}
	resp, err := client.Do(req)
	if err != nil {
		statusError.ImageRegistryNetworkStatus = err
		fmt.Printf("\n[%s] This is a Go HTTP client.Do error for: %s, %v", time.Now().Format("2006-01-02 15:04:05"), imageRegistryCheck, err)
		return // Exit the function if request fails
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err := fmt.Errorf("network error connecting to %s, status code: %d", imageRegistryCheck, resp.StatusCode)
		statusError.ImageRegistryNetworkStatus = err
		fmt.Printf("\n[%s] %v", time.Now().Format("2006-01-02 15:04:05"), err)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("\nError reading response body:", err)
		} else {
			fmt.Println("\nResponse body:", string(body))
		}
		return // Exit the function if status code is not 200
	}

	fmt.Printf("\n[%s] Network check passed for: %s, with status %v", time.Now().Format("2006-01-02 15:04:05"), imageRegistryCheck, resp.StatusCode)
}

func (statusError *StatusError) networkCheckThirdParty() {

	thirdPartyNetworkCheck := []string{"https://pypi.org/", "https://storage.googleapis.com", "https://hub.docker.com", "https://index.docker.io"}
	client := &http.Client{}

	for _, url := range thirdPartyNetworkCheck {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client req error for: %v\n", url)
			statusError.ThirdPartyNetworkStatus = append(statusError.ThirdPartyNetworkStatus, map[string]error{statement: err})
			fmt.Printf("\n[%s] %v", time.Now().Format("2006-01-02 15:04:05"), err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			statement := fmt.Sprintf("This is a Go HTTP client.Do error for: %v\n", url)
			statusError.ThirdPartyNetworkStatus = append(statusError.ThirdPartyNetworkStatus, map[string]error{statement: err})
			fmt.Printf("\n[%s] %v", time.Now().Format("2006-01-02 15:04:05"), err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err := fmt.Errorf("network error, status code: %d", resp.StatusCode)
			statusError.ThirdPartyNetworkStatus = append(statusError.ThirdPartyNetworkStatus, map[string]error{url: err})
			fmt.Printf("\n[%s] %v", time.Now().Format("2006-01-02 15:04:05"), err)
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("\nError reading response body:", err)
			} else {
				fmt.Println("\nResponse body:", string(body))
			}
			continue
		}
		fmt.Printf("\n[%s], Network check passed for: %s, with status %v\n", time.Now().Format("2006-01-02 15:04:05"), url, resp.StatusCode)
	}
}
