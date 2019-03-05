package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var params authParams
type authParams struct {
	client_id     string
	client_secret string
	response_type string
	scope         string
	redirect_uri  string
	refresh_token string
}

var accessTokenInfo ResponseInfo
type ResponseInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	UserId       string `json:"user_id"`
}

func welcomeMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	fmt.Fprintf(w, "Welcome to the Phewstoc fitbit companion app!")
}

func register(w http.ResponseWriter, r *http.Request) {
	var registerUrl = "https://www.fitbit.com/oauth2/authorize?response_type=code&client_id=" + params.client_id + "&scope=heartrate"
	http.Redirect(w, r, registerUrl, http.StatusSeeOther)
}

func concAuth(clientId string, clientSecret string) string {
	idAndSecret := clientId + ":" + clientSecret
	return b64.StdEncoding.EncodeToString([]byte(idAndSecret))
}

type HeartRate int64
type Calorie float64

type Heart struct {
	ActivitiesHeartIntraday ActivityHeartIntraday `json:"activities-heart-intraday"`
	ActivitiesHeart 				ActivitiesHeart				`json:"activities-heart"`
}

type ActivitiesHeart struct {
	Dataset  []RestingHeartRate `json:"dataset"`
}

type RestingHeartRate struct {
	HeartRate HeartRate `json:"restingHeartRate"`
}

type ActivityHeartIntraday struct {
	Dataset  []HeartIntradayDatapoint `json:"dataset"`
	Interval int                      `json:"datasetInterval"`
	Type     string                   `json:"datasetType"`
}

type HeartIntradayDatapoint struct {
	Time      string    `json:"time"`
	HeartRate HeartRate `json:"value"`
}


func authOnSuccess(w http.ResponseWriter, r *http.Request) {
	authURL := "https://api.fitbit.com/oauth2/token"

	keys, ok := r.URL.Query()["code"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url param 'code' is missing")
	}

	client := &http.Client{}

	v := url.Values{}
    v.Add("client_id", params.client_id)
	v.Add("grant_type", "authorization_code")
	v.Add("code", keys[0])

    req, err := http.NewRequest("POST", authURL, strings.NewReader(v.Encode()))
	if err != nil {
		fmt.Println("authentication error while obtaining acesstoken")
	}

	req.Header.Set("Authorization", "Basic "+concAuth(params.client_id, params.client_secret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		fmt.Println("some client Do err")
	}

	fmt.Println(resp.Status)

	err = json.NewDecoder(resp.Body).Decode(&accessTokenInfo)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		authParams.refresh_token = accessTokenInfo.RefreshToken
	}
	//fmt.Println(accessTokenInfo)
	heartRate := getHeartRateData()
	fmt.Fprint(w, heartRate.Heart)
	//fmt.Fprint(w, heartRate.ActivitiesHeartIntraday)
	//fmt.Fprint(w, heartRate.ActivitiesHeart)
}

func getHeartRateData() Heart {
	reqURL := "https://api.fitbit.com/1/user/" + accessTokenInfo.UserId + "/activities/heart/date/today/1d/1sec/time/05:00/05:01.json"
	getReq, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		fmt.Println("some get reqq err")
	}
	getReq.Header.Set("Authorization", "Bearer "+accessTokenInfo.AccessToken)
	fmt.Println(getReq)
	getResp, err := client.Do(getReq)
	if err != nil {
		fmt.Println("some client do get req error")
	}
	fmt.Println(getResp.Status)

	var heartRate Heart
	err = json.NewDecoder(getResp.Body).Decode(&heartRate)
	return heartRate
}

func fitbitData(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func refreshToken(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	v := url.Values{}
	v.Add("grant_type", "authorization_code")
	v.Add("refresh_token", authParams.refresh_token)

	req, err := http.NewRequest("POST", "https://api.fitbit.com/oauth2/token", strings.NewReader(v.Encode()))
	if err != nill {
		fmt.Println("Unable to create request.")
	}

	req.Header.Set("Authorization", "Basic "+concAuth(params.client_id, params.client_secret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		fmt.Println("some client Do err")
	}
}

func isSleeping(w http.ResponseWriter, r *http.Request) {
	// Call on register to retrive data from FitBit

	lowerbpm := findLowerHeartBeat()
	upperbpm := findUpperHeartBeat()
	timespan := findTimespanHeartBeat()

	if upperbpm - upperbpm <= 10 && timespan <= 10 {
		// return true
	} else {
		// return false
	}
}

func main() {
	params = authParams{os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]}

	http.HandleFunc("/", welcomeMessage)
	http.HandleFunc("/register/", register)
	http.HandleFunc("/success/", authOnSuccess)
	http.HandleFunc("/subscribe/sleep/", fitbitData)
	http.HandleFunc("/refresh_token/", refreshToken)
	http.HandleFunc("/is_sleeping/", isSleeping)

	err := http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/phewstoc.sladic.se/fullchain.pem", "/etc/letsencrypt/live/phewstoc.sladic.se/privkey.pem", nil)
	if err != nil {
		log.Fatal("listenAndServe: ", err)
	}
}
