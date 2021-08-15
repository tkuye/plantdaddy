package main

import (
	"time"
)

type Login struct {
	DeviceID string `json:"deviceID"`
}



type Session struct {
	SessionID string `json:"sessionID"`
	UsageCounter int `json:"usageCounter"`
	Timestamp time.Time `json:"timestamp"`
}


type SessionData struct {
	SessionID string `json:"sessionID"`
	UsageCounter int `json:"usageCounter"`
	Timestamp time.Time `json:"timestamp"`
	Temperature int `json:"temperature"`
	Humidity int `json:"humidity"`
	SoilMoisture float64 `json:"soilMoisture"`
	Light float64 `json:"light"`
}	