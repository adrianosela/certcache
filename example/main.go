package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/adrianosela/autocertDB/certcache"
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

func main() {
	hostnames := []string{
		/* YOUR DOMAINS HERE */
	}

	server := getServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("server up and running!"))
	}))

	first := autocert.DirCache(".")
	second := autocert.DirCache(".")
	third :=


	cache := certcache.NewLayered(autocert.DirCache("."), )
	certMgr := getCertManager(cache, hostnames...)

	startServer(server, certMgr)
}
