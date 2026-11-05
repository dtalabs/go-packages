package adapters

import (
	"errors"
	"net/url"
	"strings"
)

type storageConfig struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Bucket    string
	Prefix    string
	Region    string
}

func parseStorageConfig(storageURL, region string) (storageConfig, error) {
	settings := storageConfig{Region: region}
	if storageURL == "" {
		return settings, nil
	}

	parsed, err := url.Parse(storageURL)
	if err != nil {
		// Parser errors may expose credentials from the URL.
		return storageConfig{}, errors.New("invalid STORAGE_URL")
	}
	if parsed.Scheme == "s3" && parsed.User != nil {
		return parseCredentialURL(parsed, region)
	}
	if parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return storageConfig{}, errors.New("invalid STORAGE_URL")
	}

	switch parsed.Scheme {
	case "s3":
		settings.Bucket = parsed.Host
		settings.Prefix = strings.Trim(parsed.Path, "/")
	case "http", "https":
		settings.Bucket, settings.Prefix, _ = strings.Cut(strings.Trim(parsed.Path, "/"), "/")
		if settings.Bucket == "" {
			return storageConfig{}, errors.New("STORAGE_URL must include a bucket after the endpoint")
		}
		settings.Prefix = strings.Trim(settings.Prefix, "/")
		settings.Endpoint = parsed.Scheme + "://" + parsed.Host
	default:
		return storageConfig{}, errors.New("STORAGE_URL must use s3://bucket/prefix, http://endpoint/bucket/prefix, or https://endpoint/bucket/prefix")
	}
	return settings, nil
}

func parseCredentialURL(parsed *url.URL, region string) (storageConfig, error) {
	invalid := errors.New("STORAGE_URL must be s3://<key>:<secret>@<optional-endpoint>/<bucket>?region=<region> (optional scheme=http or https)")
	if parsed.Opaque != "" || parsed.Fragment != "" {
		return storageConfig{}, invalid
	}
	secret, ok := parsed.User.Password()
	if !ok || secret == "" || parsed.User.Username() == "" {
		return storageConfig{}, invalid
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return storageConfig{}, invalid
	}
	for key, values := range query {
		if (key != "region" && key != "scheme") || len(values) != 1 || strings.TrimSpace(values[0]) == "" {
			return storageConfig{}, invalid
		}
	}
	scheme := query.Get("scheme")
	if scheme == "" {
		scheme = "https"
	}
	if scheme != "http" && scheme != "https" {
		return storageConfig{}, invalid
	}
	if !strings.HasPrefix(parsed.Path, "/") {
		return storageConfig{}, invalid
	}
	bucket, prefix, _ := strings.Cut(strings.TrimPrefix(parsed.Path, "/"), "/")
	if bucket == "" || strings.ContainsAny(bucket, " \\?#@:") {
		return storageConfig{}, invalid
	}
	if query.Has("region") {
		region = query.Get("region")
	}
	if region == "" {
		region = "auto"
	}
	settings := storageConfig{
		AccessKey: parsed.User.Username(),
		SecretKey: secret,
		Bucket:    bucket,
		Prefix:    strings.Trim(prefix, "/"),
		Region:    region,
	}
	if parsed.Host != "" {
		if parsed.Hostname() == "" {
			return storageConfig{}, invalid
		}
		settings.Endpoint = scheme + "://" + parsed.Host
	} else if query.Has("scheme") {
		return storageConfig{}, invalid
	}
	return settings, nil
}
