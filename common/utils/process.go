package utils

import (
	"net"
	"strconv"
)

// ProcessAddr 将h1, h2, h3, h4, p1, p2转化为有效的TCP地址, 并返回, 为空说明参数有误
func ProcessAddr(params []string) *net.TCPAddr {
	nums := make([]int, 6)
	for i, param := range params {
		num, err := strconv.Atoi(param)
		if err != nil || num < 0 || num >= 256 {
			return nil
		}
		nums[i] = num
	}
	return &net.TCPAddr{
		IP:   net.IP{byte(nums[0]), byte(nums[1]), byte(nums[2]), byte(nums[3])},
		Port: nums[4]<<8 + nums[5],
	}
}
