package db

import "time"

type Detection struct {
	Id                uint64    `json:"id"`
	DetectionTime     time.Time `json:"detectionTime"`
	Mac               string    `json:"mac"`
	Rssi              int       `json:"rssi"`
	UptimeMs          uint64    `json:"uptimeMs"`
	BatteryPercentage uint8     `json:"batteryPercentage"`
}
