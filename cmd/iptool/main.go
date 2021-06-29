package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/bits"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/lixiangzhong/ipnet"
	"github.com/urfave/cli/v2"
)

const UsageText = `
iptool 1.0.0.0
iptool 16777216
iptool 1.0.0.0/24
iptool 1.0.0.0 1.1.0.0
iptool 1.0.0.1-255
iptool 1.0.0.0-1.0.0.255
iptool 1.0.0.0 255.255.255.0
`

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
		Name:      "iptool",
		Usage:     "方便IP计算相关的工具",
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
				for _, f := range []FuncStringBool{CIDR2IPRange, IPRange, IPInt} {
					if f(c.Args().Get(0)) {
						break
					}
				}
			case 2:
				for _, f := range []FuncStringStringBool{IPMask, IPRange2} {
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

func CIDR2IPRange(s string) (abort bool) {
	cidr, err := ipnet.ParseCIDR(s)
	if err != nil {
		return
	}
	start, end := cidr.StartEndIP()
	fmt.Println("To IP Range:")
	fmt.Println(start, end)
	_, mask := cidr.IPMask()
	fmt.Println("Mask:", mask)
	return true
}

func IPInt(s string) (abort bool) {
	{ //to int
		ip := net.ParseIP(s)
		if ip.To4() != nil {
			i := binary.BigEndian.Uint32(ip.To4())
			fmt.Println("To Uint32 BigEndian:\t", i)
			i = binary.LittleEndian.Uint32(ip.To4())
			fmt.Println("To Uint32 LittleEndian:\t", i)
			abort = true
		}
	}
	{ //to ip
		ip := make(net.IP, 4)
		i, err := strconv.ParseUint(s, 10, 32)
		if err == nil {
			binary.BigEndian.PutUint32(ip, uint32(i))
			fmt.Println("To IPv4 BigEndian:\t", ip)
			binary.LittleEndian.PutUint32(ip, uint32(i))
			fmt.Println("To IPv4 LittleEndian:\t", ip)
			abort = true
		}
	}
	return
}

func IPRange(s string) (abort bool) {
	debug(s)
	slice := strings.Split(strings.TrimSpace(s), "-")
	if len(slice) != 2 {
		return
	}
	field1 := slice[0]
	field2 := slice[1]
	startip, err := ipnet.ParseIPv4(field1)
	if err != nil {
		return
	}
	debug(field1, field2)
	var endip ipnet.IPv4
	if d, err := strconv.ParseUint(field2, 10, 8); err == nil {
		debug(d)
		if startip.IP[3] >= byte(d) {
			return
		}
		endip = ipnet.MustParseIPv4(field1)
		debug(endip)
		endip.SetD(byte(d))
		debug(endip)
	} else {
		debug(err)
		endip, err = ipnet.ParseIPv4(field2)
		if err != nil {
			return
		}
	}
	if startip.Int() > endip.Int() {
		return
	}
	cidr, err := ipnet.IPRangeToCIDR(startip.String(), endip.String())
	if err != nil {
		return
	}
	fmt.Println("To CIDR:")
	for _, v := range cidr {
		fmt.Println(v)
	}
	return true
}

func IPRange2(a, b string) bool {
	return IPRange(a + "-" + b)
}

func IPMask(ip, mask string) (abort bool) {
	cidr, err := ipnet.IPMaskToCIDR(ip, mask)
	if err != nil {
		return
	}
	if ipmask := net.ParseIP(mask).To4(); ipmask != nil {
		ones, _ := net.IPMask(ipmask).Size()
		if ones != bits.OnesCount32(binary.BigEndian.Uint32(ipmask)) {
			return
		}
	}
	fmt.Println("IPMask To CIDR:")
	fmt.Println(cidr)
	return true
}
