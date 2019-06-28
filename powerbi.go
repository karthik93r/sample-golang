package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/tkanos/gonfig"
)

type Configuration struct {
	TenentID     string
	GrantType    string
	Email        string
	Password     string
	Resource     string
	ClientId     string
	ClientSecret string
}

type PowerBiReports struct {
	AccessToken string        `json:"accessToken"`
	ExpiresOn   int64         `json:"expiresOn"`
	Reports     []interface{} `json:"reports"`
	Dashboards  []interface{} `json:"dashboards"`
}

var AccessToken string = ""
var ExpiresOn int64 = 0

func main() {
	f, err := os.OpenFile("/build_logs/testlogfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("This is a test log entry")

	for {
		time.Sleep(30 * time.Second)
		fmt.Println("Tries to get Reports")
		getAllReportsAndDashboards()
	}

}

func getAllReportsAndDashboards() {
	var respObj map[string]interface{}
	var res string
	currentTimeInMillis := time.Now().Unix()

	if int64(ExpiresOn)-currentTimeInMillis <= 100 || AccessToken == "" {
		AccessToken = ""
		respObj, res = getToken()
		if res != "" {
			fmt.Printf(res)
			log.Println(res)
			os.Exit(1)
		}
		AccessToken = respObj["access_token"].(string)
		ExpiresOn, _ = strconv.ParseInt(respObj["expires_on"].(string), 10, 64)
	}

	if AccessToken == "" {
		fmt.Printf("Error while getting access token for power bi reports")
		log.Println("Error while getting access token for power bi reports")
		os.Exit(1)
	}

	powerBiReportUrl := "https://api.powerbi.com/v1.0/myorg/reports"
	respObj, res = getDataFromPowerBI(powerBiReportUrl)

	if res != "" {
		fmt.Printf(res)
		log.Println(res)
		os.Exit(1)
	}
	reports := respObj["value"].([]interface{})

	powerBiDashboardUrl := "https://api.powerbi.com/v1.0/myorg/dashboards"
	respObj, res = getDataFromPowerBI(powerBiDashboardUrl)
	if res != "" {
		fmt.Printf(res)
		log.Println(res)
		os.Exit(1)
	}
	dashboards := respObj["value"].([]interface{})

	var obj *PowerBiReports = new(PowerBiReports)
	// obj.AccessToken = AccessToken
	obj.ExpiresOn = ExpiresOn
	obj.Reports = reports
	obj.Dashboards = dashboards

	str := fmt.Sprintf("Result: %+v", obj)
	fmt.Println(str)
	log.Println(str)
}

func getToken() (map[string]interface{}, string) {
	configuration := Configuration{}
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		return nil, fmt.Sprintf("Error in getting config.json %v:", err.Error())
	}

	oauth2TokenUrl := "https://login.microsoftonline.com/" + configuration.TenentID + "/oauth2/token"
	data := url.Values{}
	data.Set("grant_type", configuration.GrantType)
	data.Add("username", configuration.Email)
	data.Add("password", configuration.Password)
	data.Add("resource", configuration.Resource)
	data.Add("client_secret", configuration.ClientSecret)
	data.Add("client_id", configuration.ClientId)
	client := &http.Client{}
	req, err := http.NewRequest("POST", oauth2TokenUrl, bytes.NewBufferString(data.Encode()))

	if err != nil {
		return nil, fmt.Sprintf("Error in http.NewRequest for url %v: %v", oauth2TokenUrl, err.Error())
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cache-Control", "no-cache")
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Sprintf("Error in client.Do: for url %v error: %v",
			oauth2TokenUrl, err.Error())
	}

	respObj, res := processResponse(resp, http.StatusOK)

	if res != "" {
		return respObj, fmt.Sprintf(res)
	}
	return respObj, ""
}

func processResponse(resp *http.Response, expStatus int) (map[string]interface{}, string) {
	var oauth2Resp map[string]interface{} = nil
	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Sprintf("Error in ioutil.ReadAll error: %v",
			err.Error())
	}

	respStr := string(respBody)
	objBytes := []byte(respStr)
	parseErr := json.Unmarshal(objBytes, &oauth2Resp)

	if nil != parseErr {
		return nil, fmt.Sprintf("Error while unmarshalling response from oauth2 : %v", parseErr.Error())
	}

	if resp.StatusCode != expStatus {
		return nil, fmt.Sprintf("Error from JWT POST request response: %v",
			respStr)
	}
	return oauth2Resp, ""
}

func getDataFromPowerBI(powerBiUrl string) (map[string]interface{}, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", powerBiUrl, nil)
	if err != nil {
		return nil, fmt.Sprintf("Error in http.NewRequest for url %v: %v",
			powerBiUrl, err.Error())
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Sprintf("Error in client.Do: for url %v error: %v expiry: %v systime: %v",
			powerBiUrl, err.Error(), ExpiresOn, time.Now().Unix())
	}

	respObj, res := processResponse(resp, http.StatusOK)

	if res != "" {
		return nil, fmt.Sprintf(res)
	}
	return respObj, ""
}
