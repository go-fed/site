package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/go-fed/site/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type CommandLineFlags struct {
	CertFile *string
	KeyFile  *string
}

func NewCommandLineFlags() *CommandLineFlags {
	c := &CommandLineFlags{
		CertFile: flag.String("cert", "tls.crt", "Path to certificate public file"),
		KeyFile:  flag.String("key", "tls.key", "Path to certificate private key file"),
	}
	flag.Parse()
	if err := c.validate(); err != nil {
		panic(err)
	}
	return c
}

func (c *CommandLineFlags) validate() error {
	if len(*c.CertFile) == 0 {
		return fmt.Errorf("CertFile invalid: %s", *c.CertFile)
	}
	if len(*c.KeyFile) == 0 {
		return fmt.Errorf("KeyFile invalid: %s", *c.KeyFile)
	}
	return nil
}

func main() {
	c := NewCommandLineFlags()
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	httpsServer := &http.Server{
		Addr:         ":https",
		TLSConfig:    tlsConfig,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	fav := server.FaviconOptions{
		Png16Path:  "./gofed-16.png",
		Png32Path:  "./gofed-32.png",
		Png48Path:  "./gofed-48.png",
		Png96Path:  "./gofed-96.png",
		Png192Path: "./gofed-192.png",
	}
	opts := server.ServerOptions{
		TemplateFiles: []string{"tmpl.tmpl"},
		HttpServer:    httpsServer,
		RefreshRate:   time.Hour * 24,
		Favicon:       fav,
		SiteTitle:     "Go-Fed",
		OrgDataPath:   "https://github.com/go-fed",
		OrgDataName:   "GitHub",
	}
	srv, err := server.NewServer(opts)
	if err != nil {
		panic(err)
	}
	redir := &http.Server{
		Addr:         ":http",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			http.Redirect(w, req, fmt.Sprintf("https://%s%s", req.Host, req.URL), http.StatusMovedPermanently)
		}),
	}
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := redir.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP redirect server Shutdown: %v", err)
		}
		if err := httpsServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()
	go func() {
		if err := redir.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP redirect server ListenAndServe: %v", err)
		}
	}()
	if err := srv.ListenAndServeTLS(*c.CertFile, *c.KeyFile); err != nil {
		panic(err)
	}
}
