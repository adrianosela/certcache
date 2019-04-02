# SSL Certificates Storage

[![Go Report Card](https://goreportcard.com/badge/github.com/adrianosela/certcache)](https://goreportcard.com/report/github.com/adrianosela/certcache)
[![Documentation](https://godoc.org/github.com/adrianosela/certcache?status.svg)](https://godoc.org/github.com/adrianosela/certcache)
[![GitHub issues](https://img.shields.io/github/issues/adrianosela/certcache.svg)](https://github.com/adrianosela/certcache/issues)
[![license](https://img.shields.io/github/license/yangwenmai/how-to-add-badge-in-github-readme.svg)](https://github.com/adrianosela/certcache/blob/master/LICENSE)
</center>

### Go tools for managing SSL certificates from acme/autocert

The [autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) package provides automatic access to certificates from Let's Encrypt and any other ACME-based CA. This repository contains a collection of tools to simplify the task of managing certificates acquired through this method.

## Tools:
* [LayeredCache](https://godoc.org/github.com/adrianosela/certcache#LayeredCache) - chain autocert.Cache implementations
* [Functional](https://godoc.org/github.com/adrianosela/certcache#Functional) - define an autocert.Cache by using anonymous functions

## Cache Implementations:
*  [Firestore](https://godoc.org/github.com/adrianosela/certcache#Firestore) - when you want quick and easy
*  [Dynamo](https://godoc.org/github.com/adrianosela/certcache) - when your infrastructure lives in AWS **(coming soon)**
*  [Mongo](https://godoc.org/github.com/adrianosela/certcache) - for more robust applications **(coming soon)**

---

#### Why should I use this? Is this for me?

The [default storage](https://godoc.org/golang.org/x/crypto/acme/autocert#DirCache) mechanism used by autocert is the file system. Containerized and virtual workloads often don't have a persistent file system. Furthermore, file system storage is not suitable for servers spanning multiple machines or distributed systems.

See that the [autocert.Cache](https://godoc.org/golang.org/x/crypto/acme/autocert#Cache) interface is what controlls where/how certificates are stored/fetched from: 

```
m := autocert.Manager{
	Prompt:     autocert.AcceptTOS, // To always accept the terms, the callers can use AcceptTOS
	HostPolicy: autocert.HostWhitelist(hostnames...), // Specifies which hostnames the Manager is allowed to respond to
	Cache:      cache, // Cache is used by Manager to store and retrieve previously obtained certificates and other account data as opaque blobs
}
```

I have implemented the [autocert.Cache](https://godoc.org/golang.org/x/crypto/acme/autocert#Cache) interface with popular data stores on major cloud providers; so that **you dont have to!**

#### But wait, why can't I just get a new certificate every time I deploy?

Unless you have a corporate deal with Lets Encrypt, you are [limited](https://letsencrypt.org/docs/rate-limits/) to 5 duplicate certificates (certificates for the same set of names) per week on a rolling basis. This means that if your deployments don't have persistent storage, you can only deploy 5 different times (or even less if your deployments span multiple machines) within a week!

