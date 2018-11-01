package main

// Config structs contains the arguments given by go-flags from the command line.
type Config struct {
	RefreshInterval int    `long:"refresh-interval" default:"10"`
	AlertInterval   int    `long:"alert-interval" default:"120"`
	AlertThreshold  int    `long:"alert-threshold" default:"400"`
	LogInterval     int    `long:"log-interval" default:"500"`
	LogFile         string `long:"log-file" default:"/var/log/nginx/access.log"`
}

var config Config
