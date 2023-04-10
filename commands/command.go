package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/config"
	"github.com/uhppoted/uhppoted-lib/lockfile"

	"github.com/uhppoted/uhppoted-app-db/log"
)

const APP = "uhppoted-app-db"

type Options struct {
	Config string
	Debug  bool
}

type command struct {
	name        string
	description string
	usage       string

	dsn      string
	withPIN  bool
	lockfile string
	config   string
	debug    bool
}

func (cmd command) Name() string {
	return cmd.name
}

func (cmd command) Description() string {
	return cmd.description
}

func (cmd command) Usage() string {
	return cmd.usage
}

func lock(file string) (lockfile.Lockfile, error) {
	lockFile := config.Lockfile{
		File:   filepath.Join(os.TempDir(), "uhppoted-app-db.lock"),
		Remove: lockfile.RemoveLockfile,
	}

	if file != "" {
		lockFile = config.Lockfile{
			File:   file,
			Remove: lockfile.RemoveLockfile,
		}
	}

	return lockfile.MakeLockFile(lockFile)
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

func write(file string, bytes []byte) error {
	tmp, err := os.CreateTemp(os.TempDir(), "ACL")
	if err != nil {
		return err
	}

	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	fmt.Fprintf(tmp, "%s", string(bytes))
	tmp.Close()

	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0770); err != nil {
		return err
	} else if err := os.Rename(tmp.Name(), file); err != nil {
		return err
	}

	return nil
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

func errorf(tag, format string, args ...any) {
	f := fmt.Sprintf("%-10v %v", tag, format)

	log.Errorf(f, args...)
}
