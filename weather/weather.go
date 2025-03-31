package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var weatherUrl = "https://devapi.qweather.com/v7"
var client = &http.Client{
	Timeout: 50 * time.Second,
}
var key string

func init() {
	bts, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var config map[string]any
	err = json.Unmarshal(bts, &config)
	if err != nil {
		panic(err)
	}
	key = config["key"].(string)
}

// 经度 纬度
func GetNowWeather(location string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/weather/now?location=%s&key=%s", weatherUrl, location, key), nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bts, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bts, nil
}
