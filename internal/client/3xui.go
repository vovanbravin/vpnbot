package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/google/uuid"
)

type XUIClient struct {
	baseUrl     string
	client      *http.Client
	lastLogin   time.Time
	timeSession time.Duration
	username    string
	password    string
	Addr        string
	Pbk         string
	Sni         string
}

type Config struct {
	BaseUrl     string
	TimeSession time.Duration
	Username    string
	Password    string
	Addr        string
	Pbk         string
	Sni         string
	Timeout     time.Duration
}

func NewClient(config Config) (*XUIClient, error) {

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &XUIClient{
		baseUrl: config.BaseUrl,
		client: &http.Client{
			Timeout: config.Timeout * time.Second,
			Jar:     jar,
		},
		timeSession: config.TimeSession,
		username:    config.Username,
		password:    config.Password,
		Addr:        config.Addr,
		Pbk:         config.Pbk,
		Sni:         config.Sni,
	}, nil
}

func (x *XUIClient) login() error {
	url := x.baseUrl + "/login"

	body := map[string]string{
		"username": x.username,
		"password": x.password,
	}

	bodyJson, err := json.Marshal(body)

	if err != nil {
		return err
	}

	response, err := x.client.Post(url, "application/json", bytes.NewBuffer(bodyJson))

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed login")
	}

	x.lastLogin = time.Now()

	return nil
}

func (x *XUIClient) addUser(username string, id int) (string, error) {

	if time.Now().After(x.lastLogin.Add(x.timeSession)) || x.lastLogin.IsZero() {
		err := x.login()

		if err != nil {
			return "", err
		}
	}

	url := x.baseUrl + "/panel/api/inbounds/addClient"

	userUUID := uuid.New().String()

	clients := []map[string]interface{}{
		{
			"id":         userUUID,
			"flow":       "xtls-rprx-vision",
			"email":      username,
			"limitIp":    1,
			"totalGB":    0,
			"expiryTime": 0,
			"enable":     true,
		},
	}

	settingsObj := map[string]interface{}{
		"clients": clients,
	}

	settings, err := json.Marshal(settingsObj)

	if err != nil {
		return "", err
	}

	body := map[string]interface{}{
		"id":       id,
		"settings": string(settings),
	}

	bodyJson, err := json.Marshal(body)

	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJson))

	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	response, err := x.client.Do(req)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	var result map[string]interface{}

	err = json.Unmarshal(data, &result)

	if err != nil {
		return "", err
	}

	if result["success"] != true {
		return "", fmt.Errorf(result["msg"].(string))
	}

	link := fmt.Sprintf("vless://%s@%s/?type=tcp&security=reality&pbk=%s&fp=chrome&sni=%s&sid=d254c83d&spx=/&flow=xtls-rprx-vision#%s", userUUID, x.Addr, x.Pbk, x.Sni, username)

	return link, err
}
