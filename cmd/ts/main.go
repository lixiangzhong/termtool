package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
)

const UsageText = `
ts 123456789
ts 2021-06-29 18:00:00
`

const TimeFormatLayout = "2006-01-02 15:04:05"

var EnableDebug bool

func debug(v ...interface{}) {
	if EnableDebug {
		log.Println(v...)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("Debug")
	app := &cli.App{
		Name:      "ts",
		Usage:     "unix timestamp 转换",
		UsageText: UsageText,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Destination: &EnableDebug,
			},
		},
		Action: func(c *cli.Context) error {
			switch c.NArg() {
			case 1:
				for _, f := range []FuncStringBool{ParseUnix} {
					if f(c.Args().Get(0)) {
						break
					}
				}
			case 2:
				for _, f := range []FuncStringStringBool{FormatTime} {
					if f(c.Args().Get(0), c.Args().Get(1)) {
						break
					}
				}
			default:
				cli.ShowAppHelp(c)
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type FuncStringBool func(string) bool
type FuncStringStringBool func(string, string) bool

func ParseUnix(s string) (abort bool) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return
	}
	fmt.Println(time.Unix(i, 0).Format(TimeFormatLayout))
	return true
}

func FormatTime(a, b string) (abort bool) {
	t, err := time.Parse(TimeFormatLayout, a+" "+b)
	if err != nil {
		return
	}
	fmt.Println(t.Unix())
	return true
}
