package location

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var SERVER = "https://geoapi.qweather.com/v2/city/lookup"
var key string
var client = &http.Client{
	Timeout: 50 * time.Second,
}

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
func GetLocation(location string) (bts []byte, err error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?location=%s&key=%s", SERVER, location, key), nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	bts, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}
