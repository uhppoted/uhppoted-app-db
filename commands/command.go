package commands

import (
	"flag"
	"fmt"
	"time"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/config"

	"github.com/uhppoted/uhppoted-app-db/log"
)

const APP = "uhppoted-app-db"

type Options struct {
	Config string
	Debug  bool
}

func getDevices(conf *config.Config, debug bool) (uhppote.IUHPPOTE, []uhppote.Device) {
	bind, broadcast, listen := config.DefaultIpAddresses()

	if conf.BindAddress != nil {
		bind = *conf.BindAddress
	}

	if conf.BroadcastAddress != nil {
		broadcast = *conf.BroadcastAddress
	}

	if conf.ListenAddress != nil {
		listen = *conf.ListenAddress
	}

	devices := []uhppote.Device{}
	for s, d := range conf.Devices {
		// ... because d is *Device and all devices end up with the same info if you don't make a manual copy
		name := d.Name
		deviceID := s
		address := d.Address
		doors := d.Doors

		if device := uhppote.NewDevice(name, deviceID, address, doors); device != nil {
			devices = append(devices, *device)
		}
	}

	u := uhppote.NewUHPPOTE(bind, broadcast, listen, 5*time.Second, devices, debug)

	return u, devices
}

func helpOptions(flagset *flag.FlagSet) {
	count := 0
	flag.VisitAll(func(f *flag.Flag) {
		count++
	})

	flagset.VisitAll(func(f *flag.Flag) {
		fmt.Printf("    --%-13s %s\n", f.Name, f.Usage)
	})

	if count > 0 {
		fmt.Println()
		fmt.Println("  Options:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("    --%-13s %s\n", f.Name, f.Usage)
		})
	}
}

func infof(tag string, format string, args ...any) {
	f := fmt.Sprintf("%-10v %v", tag, format)

	log.Infof(f, args...)
}

func warnf(tag string, format string, args ...any) {
	f := fmt.Sprintf("%-10v %v", tag, format)

	log.Warnf(f, args...)
}
