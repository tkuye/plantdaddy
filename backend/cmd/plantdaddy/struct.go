package main

import (
	"time"
)
type NewDevice struct {
	DeviceID string `json:"deviceID"`
	DeviceName string `json:"deviceName"`
	Username string `json:"username"`
}
type Login struct {
	DeviceID string `json:"deviceID"`
}

type Device struct {
	DeviceID string `json:"deviceID"`
	DeviceName string `json:"deviceName"`
	DeviceData Data `json:"deviceData"`

}

type UserPass struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type deviceName struct {
	DeviceName string `json:"deviceName"`
	DeviceID string `json:"deviceID"`
}

type DeviceHourData struct {
	TimePeriod int `json:"timePeriod"`
	Temperature float64 `json:"temperature"`
	Humidity float64 `json:"humidity"`
	SoilMoisture float64 `json:"soilMoisture"`
	Light float64 `json:"light"`
	DeviceNumber uint64 `json:"deviceNumber"`
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
	DeviceID string `json:"deviceID"`
}

type Data struct {
	Timestamp time.Time `json:"timestamp"`
	Temperature int `json:"temperature"`
	Humidity int `json:"humidity"`
	SoilMoisture float64 `json:"soilMoisture"`
	Light float64 `json:"light"`
}
