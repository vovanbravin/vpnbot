package utils

import (
	"fmt"
	"os"
	"strconv"
	"tgbot/internal/client"
	"time"
)

func GetIntFromEnv(key string) (int, error) {
	valueStr, exist := os.LookupEnv(key)
	if !exist {
		return 0, fmt.Errorf("Key %s not exist", key)
	}

	return strconv.Atoi(valueStr)
}

func Load3xuiClientConfig() (client.Config, error) {

	baseUrl, exist := os.LookupEnv("BASE_URL_CLIENT")

	if !exist {
		return client.Config{}, fmt.Errorf("BASE_URL_CLIENT does not exist")
	}

	timeSession, err := GetIntFromEnv("TIME_SESSION")

	if err != nil {
		return client.Config{}, err
	}

	username, exist := os.LookupEnv("USERNAME_CLIENT")

	if !exist {
		return client.Config{}, fmt.Errorf("USERNAME_CLIENT does not exist")
	}

	password, exist := os.LookupEnv("PASSWORD_CLIENT")

	if !exist {
		return client.Config{}, fmt.Errorf("PASSWORD_CLIENT does not exist")
	}

	addr, exist := os.LookupEnv("ADDR_CLIENT")

	if !exist {
		return client.Config{}, fmt.Errorf("ADDR_CLIENT does not exist")
	}

	pbk, exist := os.LookupEnv("PBK_CLIENT")

	if !exist {
		return client.Config{}, fmt.Errorf("PBK_CLIENT does not exist")
	}

	sni, exist := os.LookupEnv("SNI_CLIENT")

	if !exist {
		return client.Config{}, fmt.Errorf("SNI_CLIENT does not exist")
	}

	sid, exist := os.LookupEnv("SID_CLIENT")

	if !exist {
		return client.Config{}, fmt.Errorf("SID_CLIENT does not exist")
	}

	return client.Config{
		BaseUrl:     baseUrl,
		TimeSession: time.Duration(timeSession) * time.Second,
		Username:    username,
		Password:    password,
		Addr:        addr,
		Pbk:         pbk,
		Sni:         sni,
		Timeout:     10 * time.Second,
		Sid:         sid,
	}, nil
}
