package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/skamenetskiy/ip"
)

func main() {
	//router := gin.Default()
	srv := &http.Server{
		Addr:         getEnvString("HTTP_ADDR", ":80"),
		IdleTimeout:  getEnvDuration("HTTP_IDLE_TIMEOUT", 15),
		ReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 15),
		WriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 15),
		Handler:      http.HandlerFunc(ip.Handler),
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
