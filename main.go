package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/oschwald/geoip2-golang"
)

var (
	data *geoip2.Reader
)

func init() {
	var err error
	data, err = geoip2.Open(getEnvString("GEOLITE2_FILE", "GeoLite2-City.mmdb"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	//router := gin.Default()
	srv := &http.Server{
		Addr:         getEnvString("HTTP_ADDR", ":80"),
		IdleTimeout:  getEnvDuration("HTTP_IDLE_TIMEOUT", 15),
		ReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 15),
		WriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 15),
		Handler:      http.HandlerFunc(Handler),
	}
	//router.GET("/", handler)
	tlsCert := getEnvString("TLS_CERT", "")
	tlsKey := getEnvString("TLS_KEY", "")
	go func() {
		var err error
		log.Printf("starting service on %s", srv.Addr)
		if tlsKey != "" && tlsCert != "" {
			err = srv.ListenAndServeTLS(tlsCert, tlsKey)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil {
			log.Fatal(err)
		}
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP)
	for sig := range signals {
		log.Printf("received \"%s\" signal, shutting down user server\n", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 30)
		if err := srv.Shutdown(ctx); err != nil {
			log.Println(err)
		}
		cancel()
		return
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
	//cb := r.URL.Query().Get("callback")
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

//func handler(ctx *gin.Context) {
//	ip := net.ParseIP(ctx.ClientIP())
//	city, err := data.City(ip)
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, &errResponse{
//			Code: http.StatusInternalServerError,
//			Err:  err.Error(),
//		})
//		return
//	}
//	res := &okResponse{
//		IP:       ip.String(),
//		Timezone: city.Location.TimeZone,
//		City: val{
//			Name: city.City.Names["en"],
//		},
//		Country: val{
//			Name:    city.Country.Names["en"],
//			IsoCode: city.Country.IsoCode,
//		},
//		Location: geoPoint{
//			Latitude:  city.Location.Latitude,
//			Longitude: city.Location.Longitude,
//		},
//	}
//	cb := ctx.Query("callback")
//	if cb != "" {
//		ctx.JSONP(200, res)
//		return
//	}
//	ctx.JSON(200, res)
//}

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

func getEnvString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	v, err := strconv.Atoi(getEnvString(key, "0"))
	if err != nil {
		return time.Duration(0)
	}
	return time.Duration(v)
}
