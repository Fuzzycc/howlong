package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

type (
	sizeUnit map[string]uint8
	timeUnit map[string]uint8
)

// Size Units
const (
	SU_UNKNOWN uint8 = iota
	SU_b
	SU_B
	SU_Kb
	SU_KB
	SU_Mb
	SU_MB
	SU_Gb
	SU_GB

	SU_Default1 uint8 = SU_GB
	SU_Default2 uint8 = SU_Mb
)

// Time Units
const (
	TU_UNKNOWN uint8 = iota
	TU_s
	TU_m
	TU_h
	TU_d

	TU_Default uint8 = TU_h
)

var (
	SU sizeUnit = sizeUnit{
		"b":  SU_b,
		"B":  SU_B,
		"Kb": SU_Kb,
		"KB": SU_KB,
		"Mb": SU_Mb,
		"MB": SU_MB,
		"Gb": SU_Gb,
		"GB": SU_GB,
	}

	TU timeUnit = timeUnit{
		"s": TU_s,
		"m": TU_m,
		"h": TU_h,
		"d": TU_d,
	}

	// 8 589 934 592 = 1GB in bits
	// 18446744073709600000 = 1 float64 exact
	// 1 104 107 110 113 116 119 uint64
	// uint64 is 10.ish x 10 ^19, float32 is 10^38 and float64 is 10^308
	SUR map[uint8]float64 = map[uint8]float64{
		SU_b:  1,
		SU_B:  8,
		SU_Kb: 1024,
		SU_KB: 1024 * 8,
		SU_Mb: 1024 * 1024,
		SU_MB: 1024 * 1024 * 8,
		SU_Gb: 1024 * 1024 * 1024,
		SU_GB: 1024 * 1024 * 1024 * 8,
	}

	TUR map[uint8]uint64 = map[uint8]uint64{
		TU_s: 1,
		TU_m: 60,
		TU_h: 60 * 60,
		TU_d: 24 * 60 * 60,
	}
)

const (
	ErrArgValPresent   string = "Howlong: Arg Validation Error: No Arguments provided"
	ErrArgValUnderflow string = "Howlong: Arg Validation Error: Insufficient Arguments provided"
	ErrArgOverflow     string = "Howlong: Arg Validation Error: More than enough Arguments provided, heh"
)

func main() {
	app := &cli.App{
		Name:  "howlong",
		Usage: "Never estimate download time ever again!",
		Action: func(c *cli.Context) error {
			r := processArgs(c)
			fmt.Printf("%v", r)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func processArgs(c *cli.Context) (result string) {
	if err := checkArgNumber(c); err != nil {
		log.Fatal(err)
	}
	a1 := c.Args().First()
	a2 := c.Args().Get(1)
	var a3 string
	if c.Args().Len() == 2 {
		a3 = "d"
	} else {
		a3 = c.Args().Get(2)
	}

	df, fu := parseSize(a1)
	if fu == SU_UNKNOWN {
		fu = SU_Default1
	}

	sf, su := parseSize(a2)
	if su == SU_UNKNOWN {
		su = SU_Default2
	}
	// fmt.Println("log: sf: ", sf, " su: ", su)

	tu := parseTime(a3)
	if tu == TU_UNKNOWN {
		tu = TU_Default
	}

	db := reduceSize(df, fu)
	sb := reduceSize(sf, su)

	total := db / sb
	result = strconv.FormatUint(total / reduceTime(tu), 10)
	// fmt.Println(result == 0, result)
	if result == "0" {
		result = strconv.FormatFloat(float64(float32(total) / reduceTimeFloat(tu)), 'f', 2, 64)
	}

	return result
}

func reduceTime(u uint8) (n uint64) {
	n = 1
	if value, ok := TUR[u]; ok {
		n = value
		return
	} else {
		return
	}
}

func reduceTimeFloat(u uint8) (f float32){
	f = 1
	if value, ok := TUR[u]; ok {
		f = (float32(value))
		return
	} else {
		return
	}
}

// Reduces f by using u into equivalent n bits.
//
// Returns 0 on failure. Hopefully.
func reduceSize(f float32, u uint8) (n uint64) {
	n = 0

	t := float64(f)
	// Absolute the float to get positive number
	t = math.Abs(t)
	// trim float to 3 decimals
	t = math.Floor(t*1000+0.5) / 1000

	// use SUR by referencing u to turn it into bits
	if value, ok := SUR[u]; ok {
		t *= value
	} else {
		return
	}
	// example: f=1.1 and u=T_GB
	// example: 1*SUR[T_GB] = 9448928051.2

	// cut any decimals, since it's in bits now, it won't matter if we just cut away decimal bits
	t = math.Floor(t)
	// they are so insignificant for our purposes

	// turn it into uint64, its more than enough for our purposes. float64 is just too big
	n = uint64(t)
	// return the number
	// fmt.Println("log:", n, "u: ", u)
	return n
}

func checkArgNumber(c *cli.Context) error {
	args := c.Args()
	alen := args.Len()

	// fail switch
	switch {
	case !args.Present():
		return errors.New(ErrArgValPresent)
	case alen < 2:
		return errors.New(ErrArgValUnderflow)
	case alen > 3:
		return errors.New(ErrArgOverflow)
	default:
		return nil
	}
}

// Parse a valid s into n and u
// If s is invalid, returns n = 0 and u = [SU_UNKNOWN]
// if s is valid but lacks u, returns n and u = [SU_UNKNOWN]
func parseSize(s string) (n float32, u uint8) {
	u = SU_UNKNOWN

	// get the unit
	for key, value := range SU {
		if strings.HasSuffix(s, key) {
			if key == "b" || key == "B" {
				// skip b or B because
				// GB for example will remove B and leave G
				// breaking the code
				continue
			}
			u = value
			s = strings.TrimSuffix(s, key)
			break
		}
	}
	if strings.HasSuffix(s, "b") {
		u = SU["b"]
		s = strings.TrimSuffix(s, "b")
	} else if strings.HasSuffix(s, "B") {
		u = SU["B"]
		s = strings.TrimSuffix(s, "B")
	}

	// // fmt.Printf("s: %+v\n", s)
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		// fmt.Printf("Error parsing\n")
		return 0, u
	}
	n = float32(f)

	// fmt.Printf("log: n: %+v u: %+v\n",n, u)
	return n, u
}

// Parse a valid s into u
// if s is invalid, returns u = [TU_UNKNOWN]
func parseTime(s string) (u uint8) {
	u = TU_UNKNOWN

	if len(s) != 1 {
		return
	}

	// get the unit
	if elm, ok := TU[s]; ok {
		u = elm
		return
	} else {
		return
	}
}
