package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/lixiangzhong/termtool/pkg/netaddr"
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
				for _, f := range []FuncStringBool{CIDR2IPRange, IPRange, PrintBits, IPInt} {
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
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return
	}
	// cidr, err := ipnet.ParseCIDR(s)

	start, end := netaddr.CIDRToIPRange(p)
	// start, end := cidr.StartEndIP()
	fmt.Println("To IP Range:")
	fmt.Println(start, end)
	// _, mask := cidr.IPMask()
	mask := netaddr.CIDRNetMask(p)
	fmt.Println("Mask:", mask)
	return true
}

func IPInt(s string) (abort bool) {

	{ //to int
		addr, err := netip.ParseAddr(s)
		if err == nil {
			if addr.Is4() {
				i := binary.BigEndian.Uint32(addr.AsSlice())
				fmt.Println("To Uint32 BigEndian:\t", i)
				i = binary.LittleEndian.Uint32(addr.AsSlice())
				fmt.Println("To Uint32 LittleEndian:\t", i)
				abort = true
			}
			if addr.Is6() {
				fmt.Println("To BigInt:\t", big.NewInt(0).SetBytes(addr.AsSlice()))
				abort = true
			}
		}
	}
	{ //to ip

		ii, ok := big.NewInt(0).SetString(s, 10)
		if ok {
			ip6 := make([]byte, 16)
			ii.FillBytes(ip6)
			addr, ok := netip.AddrFromSlice(ip6)
			if ok {
				fmt.Println("To IPv6:\t", addr)
			}
		}
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
	startip, err := netip.ParseAddr(field1)
	if err != nil {
		return
	}
	debug(field1, field2)
	var endip netip.Addr
	if d, err := strconv.ParseUint(field2, 10, 8); err == nil {
		debug(d)
		// if startip.IP[3] >= byte(d) {
		// 	return
		// }
		slice := startip.AsSlice()
		if slice[len(slice)-1] >= byte(d) {
			return
		}
		slice[len(slice)-1] = byte(d)
		var ok bool
		endip, ok = netip.AddrFromSlice(slice)
		if !ok {
			return
		}
	} else {
		debug(err)
		endip, err = netip.ParseAddr(field2)
		if err != nil {
			return
		}
	}
	if startip.Compare(endip) > 0 {
		return
	}
	cidr, err := netaddr.IPRangeToCIDR(startip.String(), endip.String())
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
	a, err := netip.ParseAddr(ip)
	if err != nil {
		return
	}
	m, err := netip.ParseAddr(mask)
	if err != nil {
		return
	}
	p, err := netaddr.IPMaskToCIDR(a, m)
	if err != nil {
		return
	}
	// cidr, err := ipnet.IPMaskToCIDR(ip, mask)
	// if err != nil {
	// 	return
	// }
	// if ipmask := net.ParseIP(mask).To4(); ipmask != nil {
	// 	ones, _ := net.IPMask(ipmask).Size()
	// 	if ones != bits.OnesCount32(binary.BigEndian.Uint32(ipmask)) {
	// 		return
	// 	}
	// }
	fmt.Println("IPMask To CIDR:")
	fmt.Println(p)
	return true
}

func PrintBits(s string) (abort bool) {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return
	}
	slice := addr.AsSlice()
	length := len(slice)
	fmt.Printf("BigEndian Bits:\t\t")
	for i := 0; i < length; i++ {
		fmt.Printf("%08b ", slice[i])
		if i == length-1 {
			fmt.Println()
		}
	}
	fmt.Printf("LittleEndian Bits:\t")
	for i := 0; i < length; i++ {
		fmt.Printf("%08b ", slice[length-i-1])
		if i == length-1 {
			fmt.Println()
		}
	}
	return false
	// ip := net.ParseIP(s)
	// if ip.To4() == nil {
	// 	return
	// }
	// ip = ip.To4()
	// fmt.Printf("BigEndian Bits:\t\t %08b %08b %08b %08b\n", ip[0], ip[1], ip[2], ip[3])
	// fmt.Printf("LittleEndian Bits:\t %08b %08b %08b %08b\n", ip[3], ip[2], ip[1], ip[0])
	// return false
}
