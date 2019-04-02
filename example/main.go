package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adrianosela/certcache"
	"golang.org/x/crypto/acme/autocert"
)

func startServer(s *http.Server, mg *autocert.Manager) {
	s.Addr = ":443"
	s.TLSConfig = &tls.Config{GetCertificate: mg.GetCertificate}
	go func() {
		log.Printf("Serving HTTPS at %s", s.Addr)
		err := s.ListenAndServeTLS("", "")
		if err != nil {
			log.Fatalf("ListendAndServeTLS() failed with %s", err)
		}
	}()
	// allow autocert handler Let's Encrypt auth callbacks over HTTP
	s.Handler = mg.HTTPHandler(s.Handler)
	// some time for OS scheduler to start SSL thread
	time.Sleep(time.Millisecond * 50)

	s.Addr = ":80"
	log.Printf("Serving HTTP at %s", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("httpSrv.ListenAndServe() failed with %s", err)
	}
}

func getServer(h http.Handler) *http.Server {
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      h,
	}
}

func getCertManager(cache autocert.Cache, hostnames ...string) *autocert.Manager {
	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hostnames...),
		Cache:      cache,
	}
}

// getLoggerLayer is an example of what kind of clever things
// can be done using the certcache.Functional type
func getLoggerLayer() *certcache.Functional {
	return certcache.NewFunctional(
		func(ctx context.Context, key string) ([]byte, error) {
			log.Printf("[certcache] getting key %s", key)
			return nil, autocert.ErrCacheMiss
		},
		func(ctx context.Context, key string, data []byte) error {
			log.Printf("[certcache] putting key %s", key)
			return nil
		},
		func(ctx context.Context, key string) error {
			log.Printf("[certcache] deleting key %s", key)
			return nil
		},
	)
}

func main() {
	hostnames := []string{
		/* YOUR DOMAIN HERE (remove the example) */
		"certcachetest.adrianosela.com",
	}

	server := getServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("server up and running!"))
	}))

	firstLayer := getLoggerLayer()
	secondLayer := autocert.DirCache(".")
	thirdLayer := certcache.NewFirestore(os.Getenv("FIRESTORE_CREDS_PATH"), os.Getenv("FIRESTORE_PROJ_ID"))

	cache := certcache.NewLayered(firstLayer, secondLayer, thirdLayer)
	certMgr := getCertManager(cache, hostnames...)

	startServer(server, certMgr)
}
