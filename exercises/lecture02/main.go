package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ResponseJSON struct {
	Current CurrentJSON `json:"current"`
}

type CurrentJSON struct {
	TempC     float64       `json:"temp_c"`
	Condition ConditionJSON `json:"condition"`
}

type ConditionJSON struct {
	Text string `json:"text"`
}

type OnlineSinoptik struct {
	api string
}

func NewOnlineSinoptik() *OnlineSinoptik {
	return &OnlineSinoptik{
		api: "https://api.weatherapi.com/v1/current.json?key=<your key here>",
	}
}

func (os *OnlineSinoptik) Weather(location string) (Weather, error) {
	locationQuery := fmt.Sprintf("q=%s", location)
	resp, err := http.Get(fmt.Sprintf("%s&%s", os.api, locationQuery))
	if err != nil {
		return Weather{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Weather{}, err
	}

	var r ResponseJSON
	err = json.Unmarshal(bodyBytes, &r)
	if err != nil {
		return Weather{}, err
	}

	return Weather{
		TempC:   r.Current.TempC,
		Details: ParseWeatherDetails(r.Current.Condition.Text),
	}, nil
}

type Weather struct {
	TempC   float64
	Details WeatherDetails
}

type WeatherDetails string

func ParseWeatherDetails(s string) WeatherDetails {
	switch {
	case strings.EqualFold(s, "overcast"):
		return WeatherDetailsOvercast
	case strings.EqualFold(s, "light rain"):
		return WeatherDetailsRain
	default:
		return WeatherDetailsUnknown
	}
}

const (
	WeatherDetailsFoggy    WeatherDetails = "foggy"
	WeatherDetailsSunny    WeatherDetails = "sunny"
	WeatherDetailsOvercast WeatherDetails = "overcast"
	WeatherDetailsRain     WeatherDetails = "rain"
	WeatherDetailsUnknown  WeatherDetails = "unknown"
)

type Sinoptik struct {
	w map[string]Weather
}

func NewSinoptik() *Sinoptik {
	return &Sinoptik{
		w: map[string]Weather{
			"Sofia":  Weather{TempC: 11, Details: "foggy"},
			"London": Weather{TempC: 1, Details: "rain"},
		},
	}
}

func (s *Sinoptik) Weather(location string) Weather {
	return s.w[location]
}

type Assistant struct{}

func (a *Assistant) DoINeedAnUmbrella(w Weather) bool {
	if w.Details == WeatherDetailsOvercast || w.Details == WeatherDetailsRain {
		return true
	}
	return false
}

func main() {
	location := "Qatar"
	sinoptik := NewOnlineSinoptik()

	w, err := sinoptik.Weather(location)
	if err != nil {
		fmt.Println("something went wrong:", err)
		os.Exit(1)
	}
	fmt.Printf("Temp is %f, details: %s.\n", w.TempC, w.Details)
	a := &Assistant{}
	fmt.Printf("Umbrella? %v\n", a.DoINeedAnUmbrella(w))
}
