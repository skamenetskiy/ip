package ip

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

var (
	data *geoip2.Reader
)

func init() {
	var err error
	data, err = geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ip := net.ParseIP(getIP(r))
	log.Println("ip:", ip)
	city, err := data.City(ip)
	if err != nil {
		if err := json.NewEncoder(w).Encode(&errResponse{
			Code: http.StatusInternalServerError,
			Err:  err.Error(),
		}); err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res := &okResponse{
		IP:       ip.String(),
		Timezone: city.Location.TimeZone,
		City: val{
			Name: city.City.Names["en"],
		},
		Country: val{
			Name:    city.Country.Names["en"],
			IsoCode: city.Country.IsoCode,
		},
		Location: geoPoint{
			Latitude:  city.Location.Latitude,
			Longitude: city.Location.Longitude,
		},
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println(err)
	}
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	ip = r.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Appengine-Remote-Addr")
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

type okResponse struct {
	IP       string   `json:"ip"`
	Timezone string   `json:"timezone"`
	Country  val      `json:"country"`
	City     val      `json:"city"`
	Location geoPoint `json:"location"`
}

type errResponse struct {
	Err  string `json:"err"`
	Code int    `json:"code"`
}

type val struct {
	Name    string `json:"name"`
	IsoCode string `json:"isoCode,omitempty"`
}

type geoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
