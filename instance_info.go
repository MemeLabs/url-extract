package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// InstanceInfo is the information that is provided by the debugger instance.
// By default, the information is exposed on http://localhost:9222/json/version
// reference: https://chromium.googlesource.com/external/github.com/mafredri/cdp/+/a974e2fd933e19fc0bbde4ea092df45158e782bf
type InstanceInfo struct {
	Browser              string `json:"Browser"`
	ProtocolVersion      string `json:"Protocol-Version"`
	UserAgent            string `json:"User-Agent"`
	V8Version            string `json:"V8-Version"`
	WebKitVersion        string `json:"WebKit-Version"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// GetInstanceInfo fetches information about an instance running at the given endpoint.
// E.g. "localhost:9222".
func GetInstanceInfo(endpoint string) (*InstanceInfo, error) {

	client := &http.Client{
		Timeout: time.Second * 3,
	}

	resp, err := client.Get(fmt.Sprintf("http://%s/json/version", endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ii InstanceInfo
	err = json.Unmarshal(contents, &ii)
	if err != nil {
		return nil, err
	}
	return &ii, nil
}
