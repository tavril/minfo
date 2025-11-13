package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"
	"time"
)

var (
	appName              = path.Base(os.Args[0])
	arch                 = runtime.GOARCH
	defaultConfigFile    = fmt.Sprintf("%s/.config/%s/config.yaml", os.Getenv("HOME"), appName)
	weatherCacheFile     = fmt.Sprintf("%s/.cache/%s/weather.json", os.Getenv("HOME"), appName)
	weatherCacheDuration = 15 * time.Minute
	reANSI               = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
	envHome              = os.Getenv("HOME")
	hostInfo             = info{}
	GitCommit            string
	GitVersion           string
	colorNormal          = "\u001B[0m"
	colorCyan            string // The colors will be defined depending on the terminal type (256 or 16 colors)
)

var wmoCodesDesc = map[int]map[string]string{
	0: {
		"en": "Clear sky",
		"fr": "Dégagé",
	},
	1: {
		"en": "Mainly clear",
		"fr": "Principalement dégagé",
	},
	2: {
		"en": "Partly cloudy",
		"fr": "Partiellement nuageux",
	},
	3: {
		"en": "Overcast",
		"fr": "Couvert",
	},
	45: {
		"en": "Fog",
		"fr": "Brouillard",
	},
	48: {
		"en": "Depositing rime fog",
		"fr": "Brouillard givrant",
	},
	51: {
		"en": "Light drizzle",
		"fr": "Légère bruine",
	},
	53: {
		"en": "Drizzle",
		"fr": "Bruine",
	},
	55: {
		"en": "Dense drizzle",
		"fr": "Bruine dense",
	},
	56: {
		"en": "Light freezing drizzle",
		"fr": "Légère bruine verglaçante",
	},
	57: {
		"en": "Dense freezing Drizzle",
		"fr": "Bruine verglaçante dense",
	},
	61: {
		"en": "Slight rain",
		"fr": "Légère pluie",
	},
	63: {
		"en": "Rain",
		"fr": "Pluie",
	},
	65: {
		"en": "Heavy rain",
		"fr": "Forte pluie",
	},
	66: {
		"en": "Light freezing rain",
		"fr": "Légère pluie verglaçante",
	},
	67: {
		"en": "Heavy freezing rain",
		"fr": "Forte pluie verglaçante",
	},
	71: {
		"en": "Slight snow fall",
		"fr": "Chute de neige",
	},
	73: {
		"en": "Snow fall",
		"fr": "Chute de neige modérée",
	},
	75: {
		"en": "Heavy snow fall",
		"fr": "Forte chute de neige",
	},
	77: {
		"en": "Snow grains",
		"fr": "Neige en grains",
	},
	80: {
		"en": "Slight rain showers",
		"fr": "Légère averse de pluie",
	},
	81: {
		"en": "Rain showers",
		"fr": "Averse de pluie",
	},
	82: {
		"en": "Heavy rain showers",
		"fr": "Forte averse de pluie",
	},
	85: {
		"en": "Slight snow showers",
		"fr": "Légère averse de neige",
	},
	86: {
		"en": "Heavy snow showers",
		"fr": "Forte averse de neige",
	},
	95: {
		"en": "Thunderstorm",
		"fr": "Orageux",
	},
	96: {
		"en": "Slight thunderstorm with hail",
		"fr": "Léger orage accompagné de grêle",
	},
	99: {
		"en": "Heavy thunderstorm with hail",
		"fr": "Fort orage accompagné de grêle",
	},
}
