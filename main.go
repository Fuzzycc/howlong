package main

import (
	"bufio"
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

	// Inverse of SU, mapping uint8 to string
	I_SU map[uint8]string = map[uint8]string{
		SU_b:  "b",
		SU_B:  "B",
		SU_Kb: "Kb",
		SU_KB: "KB",
		SU_Mb: "Mb",
		SU_MB: "MB",
		SU_Gb: "Gb",
		SU_GB: "GB",
	}

	// Inverse of TU, mapping uint8 to string
	I_TU map[uint8]string = map[uint8]string{
		TU_s: "s",
		TU_m: "m",
		TU_h: "h",
		TU_d: "d",
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

const (
	Version = "v0.6.5"
)

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("%s\n", cCtx.App.Version)
	}

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`

	cli.HelpFlag = &cli.BoolFlag{
		Name:               "help",
		Aliases:            []string{"h"},
		Usage:              "Show this help text",
		DisableDefaultText: true,
	}

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "continuous",
				Aliases: []string{"c"},
				Usage:   "Enter continuous mode with the provided `SIZEUNIT SPEEDUNIT TIMEUNIT`",
			},
		},
		Name:    "howlong",
		Version: Version,
		Suggest: true,
		Authors: []*cli.Author{
			{
				Name: "Fuzzycc@github",
			},
		},
		Usage:     "Never estimate download time ever again!",
		ArgsUsage: "{size[unit]} {speed[unit]} [time-unit]",
		Action: func(ctx *cli.Context) error {
			if ctx.Bool("continuous") {
				processContinuous(ctx)
				return nil
			} else {
				r := processArgs(ctx)
				fmt.Printf("%v", r)
				return nil
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// Deprecated
// func readCont(ctx *cli.Context) {
// 	var fu, su, tu uint8 = processArgsContinuous(ctx)

// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		in := scanner.Text()
// 		in = processInput(in, fu, su, tu)
// 		fmt.Println(in + I_TU[tu])
// 		if err := scanner.Err(); err != nil {
// 			fmt.Println("howlong: error reading from stdin: ", err)
// 		}
// 	}
// }

func processContinuous(ctx *cli.Context) {
	var fu, su, tu uint8 = processArgsContinuous(ctx)

	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		in := sc.Text()
		in = processInput(in, fu, su, tu)
		fmt.Println(in + I_TU[tu])
	}

	err := sc.Err()
	var errout string
	switch err {
	case nil: // nil is EOF
		errout = "EOF"
	default:
		errout = "howlong: error reading from stdin: " + err.Error()
	}
	fmt.Println(errout)
	return
}

// Deprecated
// func wrapperContinuous(ctx *cli.Context) {
// 	sigs := make(chan os.Signal, 1)
// 	defer close(sigs)

// 	q := make(chan bool, 1)
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

// 	input := make(chan string, 1)
// 	defer close(input)

// 	var fu, su, tu uint8 = processArgsContinuous(ctx)

// 	scanner := bufio.NewReader(os.Stdin)

// 	go func() {
// 		<-sigs
// 		fmt.Println("-1")
// 		close(q)
// 	}()

// loop:
// 	for {
// 		in := readInput(scanner)
// 		if in == "" {
// 			break loop
// 		}
// 		in = processInput(in, fu, su, tu)
// 		fmt.Println(in + I_TU[tu])
// 	}
// 	select {
// 	case <-q:
// 	}
// }

// Deprecated
// func readInput(scanner *bufio.Reader) (s string) {
// 	s, err := scanner.ReadString('\n')
// 	if err == io.EOF {
// 		return ""
// 	}
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	s = strings.TrimRight(s, "\r\n")
// 	// fmt.Println("-", s, "-")
// 	return s
// 	// return scanner.Text()
// }

func processInput(in string, fu uint8, su uint8, tu uint8) (result string) {
	parts := strings.Split(in, " ")
	err := func() error {
		plen := len(parts)
		switch plen {
		// case 0: // this never happens, len is always minimum of 1, read strings.Split
		// 	return errors.New("Howlong: Input Validation Error: 0 input, 2 required")
		case 1:
			return errors.New("Howlong: Input Validation Error: 1 input, 2 required")
		case 2:
			return nil
		case 3:
			return errors.New("Howlong: Input Validation Error: >2 input, 2 required")
		default:
			return errors.New("Howlong: Input Validation Error: Invalid Input")
		}
	}()
	if err != nil {
		log.Fatal(err)
	}

	// suffixing with unit to make parseSize work
	// might do a dedicated parser if it's faster
	a1 := parts[0] + I_SU[fu]
	a2 := parts[1] + I_SU[su]

	df, _ := parseSize(a1)
	sf, _ := parseSize(a2)

	db := reduceSize(df, fu)
	sb := reduceSize(sf, su)
	if sb == 0 {
		sb = 1
	}

	total := db / sb

	result = strconv.FormatFloat(float64(float32(total)/reduceTimeFloat(tu)), 'f', 4, 64)

	return result
}

func processArgsContinuous(c *cli.Context) (fu uint8, su uint8, tu uint8) {
	// Init units from args
	// 2 or 3 arguments
	if err := checkArgNumber(c); err != nil {
		log.Fatal(err)
	}
	a1 := c.Args().First()
	a2 := c.Args().Get(1)
	var a3 string
	if c.Args().Len() == 2 {
		a3 = ""
	} else {
		a3 = c.Args().Get(2)
	}

	a1 = "10" + a1
	_, fu = parseSize(a1)
	if fu == SU_UNKNOWN {
		fu = SU_Default1
	}

	a2 = "10" + a2
	_, su = parseSize(a2)
	if su == SU_UNKNOWN {
		su = SU_Default2
	}

	tu = parseTime(a3)
	if tu == TU_UNKNOWN {
		tu = TU_Default
	}

	return fu, su, tu
}

// The main brain. This function turns args into the output string
func processArgs(c *cli.Context) (result string) {
	if err := checkArgNumber(c); err != nil {
		log.Fatal(err)
	}
	a1 := c.Args().First()
	a2 := c.Args().Get(1)
	var a3 string
	if c.Args().Len() == 2 {
		a3 = ""
	} else {
		a3 = c.Args().Get(2)
	}

	df, fu := parseSize(a1)
	if df == 0 {
		log.Fatal("Invalid size format: ", a1)
	}
	if fu == SU_UNKNOWN {
		fu = SU_Default1
	}

	sf, su := parseSize(a2)
	if sf == 0 {
		log.Fatal("Invalid size format: ", a2)
	}
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
	if sb == 0 {
		sb = 1
	}

	total := db / sb
	// result = strconv.FormatUint(total/reduceTime(tu), 10)
	// // fmt.Println(result == 0, result)
	// if result == "0" {
	// 	result = strconv.FormatFloat(float64(float32(total)/reduceTimeFloat(tu)), 'f', 4, 64)
	// }

	// all float, its better
	result = strconv.FormatFloat(float64(float32(total)/reduceTimeFloat(tu)), 'f', 4, 64) + I_TU[tu]

	return result
}

// Uses TUR to reduce a TU to a second-based u uint64
func reduceTime(u uint8) (n uint64) {
	n = 1
	if value, ok := TUR[u]; ok {
		n = value
		return
	} else {
		return
	}
}

// Uses TUR to reduce a TU to a second-based u float
func reduceTimeFloat(u uint8) (f float32) {
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
//
// If s is invalid, returns n = 0 and u = [SU_UNKNOWN]
//
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
		log.Fatal("howlong: ", err)
		// os.Exit(12)
		// return 0, u
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
