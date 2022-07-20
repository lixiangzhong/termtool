package netaddr

import (
	"fmt"
	"math/big"
	"net/netip"
	"sort"
)

func reverseBit(v *big.Int, bitLen int) *big.Int {
	tmp := big.NewInt(0).Lsh(big.NewInt(1), uint(bitLen))
	tmp = tmp.Sub(tmp, big.NewInt(1)).Xor(tmp, v)
	return tmp
}

func ipBigInt(s string) *big.Int {
	addr := netip.MustParseAddr(s)
	return big.NewInt(0).SetBytes(addr.AsSlice())
}

func lsh(i *big.Int, n uint) *big.Int {
	return big.NewInt(0).Lsh(i, n)
}

func rsh(i *big.Int, n uint) *big.Int {
	return big.NewInt(0).Rsh(i, n)
}

func IPRangeToCIDR(startip, endip string) ([]netip.Prefix, error) {
	start := ipBigInt(startip)
	end := ipBigInt(endip)
	var bitLen = 32
	if end.BitLen() > bitLen {
		bitLen = 128
	}
	var data []netip.Prefix
	var i int
	for end.Cmp(start) >= 0 {
		bit := reverseBit(end, bitLen).TrailingZeroBits()
		if i == 0 && bit == 0 || end.Cmp(start) == 0 {
			addr, ok := netip.AddrFromSlice(end.Bytes())
			if !ok {
				return data, fmt.Errorf("error:AddrFromSlice:%v", end.Bytes())
			}
			data = append(data, netip.PrefixFrom(addr, bitLen))
			end.Sub(end, big.NewInt(1))
		}
		i++
		for bit > 0 {
			begin := lsh(rsh(end, bit), bit)
			if begin.Cmp(start) < 0 {
				bit--
			} else {
				addr, ok := netip.AddrFromSlice(begin.Bytes())
				if !ok {
					return data, fmt.Errorf("error:AddrFromSlice:%v", end.Bytes())
				}
				data = append(data, netip.PrefixFrom(addr, bitLen-int(bit)))
				if start.Cmp(begin) == 0 {
					return data, nil
				}
				end.Sub(begin, big.NewInt(1))
				break
			}
		}
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Addr().Compare(data[j].Addr()) < 0
	})
	return data, nil
}

func sub(i *big.Int, n *big.Int) *big.Int {
	return big.NewInt(0).Sub(i, n)
}

func CIDRToIPRange(p netip.Prefix) (start, end netip.Addr) {
	p = p.Masked()
	start = p.Addr()

	i := big.NewInt(0).SetBytes(p.Addr().AsSlice())
	hostMask := sub(lsh(big.NewInt(1), uint(p.Addr().BitLen()-p.Bits())), big.NewInt(1))
	i.Or(i, hostMask)
	b := make([]byte, p.Addr().BitLen()/8)
	i.FillBytes(b)
	end, _ = netip.AddrFromSlice(b)
	return
}

func CIDRNetMask(p netip.Prefix) (mask netip.Addr) {
	p = p.Masked()
	i := sub(lsh(big.NewInt(1), uint(p.Bits())), big.NewInt(1))
	i = lsh(i, uint(p.Addr().BitLen()-p.Bits()))
	b := make([]byte, p.Addr().BitLen()/8)
	i.FillBytes(b)
	mask, _ = netip.AddrFromSlice(b)
	return
}

func IPMaskToCIDR(ip, mask netip.Addr) (netip.Prefix, error) {
	bits := mask.BitLen() - int(big.NewInt(0).SetBytes(mask.AsSlice()).TrailingZeroBits())
	return ip.Prefix(bits)
}
