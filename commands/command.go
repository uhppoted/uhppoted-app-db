package commands

import (
//	"flag"
//	"fmt"

// "github.com/uhppoted/uhppoted-app-db/log"
)

const APP = "uhppoted-app-db"

type Options struct {
	Config string
	Debug  bool
}

//func helpOptions(flagset *flag.FlagSet) {
//	count := 0
//	flag.VisitAll(func(f *flag.Flag) {
//		count++
//	})
//
//	flagset.VisitAll(func(f *flag.Flag) {
//		fmt.Printf("    --%-13s %s\n", f.Name, f.Usage)
//	})
//
//	if count > 0 {
//		fmt.Println()
//		fmt.Println("  Options:")
//		flag.VisitAll(func(f *flag.Flag) {
//			fmt.Printf("    --%-13s %s\n", f.Name, f.Usage)
//		})
//	}
//}

// func infof(format string, args ...any) {
// 	log.Infof(format, args...)
// }
