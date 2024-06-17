package main

import (
	"fmt"
	"strconv"
	"time"
)

type Progress struct {
	Device     string  `json:"device"`
	Progress   string  `json:"progress"`
	Document   string  `json:"document"`
	Percentage float64 `json:"percentage"`
	DeviceID   string  `json:"device_id"`
	Timestamp  int64   `json:"timestamp"`
	User       User
}

func (p *Progress) DocKey() string {
	return fmt.Sprintf("user:%s:document:%s", p.User.Username, p.Document)
}

func (p *Progress) Save(d *Store) {
	d.Set(p.DocKey()+"_percent", fmt.Sprintf("%f", p.Percentage))
	d.Set(p.DocKey()+"_progress", p.Progress)
	d.Set(p.DocKey()+"_device", p.Device)
	d.Set(p.DocKey()+"_device_id", p.DeviceID)
	d.Set(p.DocKey()+"_timestamp", fmt.Sprintf("%d", (time.Now().Unix())))
}

func (p *Progress) Get(d *Store) error {
	if p.Document == "" {
		return fmt.Errorf("invalid document")
	}

	pct, err := d.Get(p.DocKey() + "_percent")
	if err != nil {
		return err
	}
	p.Percentage, _ = strconv.ParseFloat(string(pct), 64)

	prog, err := d.Get(p.DocKey() + "_progress")
	if err != nil {
		return err
	}
	p.Progress = string(prog)

	dev, err := d.Get(p.DocKey() + "_device")
	if err != nil {
		return err
	}
	p.Device = string(dev)

	devID, err := d.Get(p.DocKey() + "_device_id")
	if err != nil {
		return err
	}
	p.DeviceID = string(devID)

	ts, err := d.Get(p.DocKey() + "_timestamp")
	if err != nil {
		return err
	}
	stamp, err := strconv.ParseInt(string(ts), 10, 64)
	if err != nil {
		return err
	}

	p.Timestamp = stamp

	return nil
}
