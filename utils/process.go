package utils

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"net"
	"strconv"
	"strings"
)

// WrapAddr 将h1, h2, h3, h4, p1, p2转化为有效的TCP地址, 并返回, 为空说明参数有误
func WrapAddr(params []string) *net.TCPAddr {
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

// FormatFileList 将文件夹的信息格式化, Unix文件列表格式
func FormatFileList(list []fs.FileInfo) string {
	builder := new(strings.Builder)
	for _, info := range list {
		builder.WriteString(info.Mode().String())
		builder.WriteString(" 1 unknown unknown ")
		builder.WriteString(strconv.FormatInt(info.Size(), 10))
		builder.WriteString(info.ModTime().Format(" Jan _2 15:04 "))
		builder.WriteString(info.Name())
		builder.WriteString("\r\n")
	}
	return builder.String()
}

// FormatAddr 把地址格式化
func FormatAddr(addr *net.TCPAddr) string {
	ip := addr.IP.To4()
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d", ip[0], ip[1], ip[2], ip[3],
		addr.Port>>8, addr.Port&0xff,
	)
}

// HashStrings 把字符串形式的密码转化为16位的二进制数串
func HashStrings(password string) []byte {
	hash := md5.New()
	hash.Write([]byte(password))
	return hash.Sum(nil)
}

// StringsIsNotNull 检查字符串是否为空, 若不为空则返回true, 否则返回false
func StringsIsNotNull(str string) bool {
	if strings.TrimSpace(str) == "" {
		return false
	}
	return true
}

// TransferPageToOffset 把页数转化为位移量, page: 页数; limit: 页内实体数量; sum: 实体总数
func TransferPageToOffset(page, limit, sum int) int {
	minPage, maxPage := CalPageExtreme(limit, sum)
	if page < minPage || page > maxPage {
		return 0
	}
	return (page - 1) * limit
}

// CalPageExtreme 根据页内实体数量和实体总数计算页数的边界
func CalPageExtreme(limit, sum int) (int, int) {
	minPage, maxPage := 1, 1
	if sum == 0 {
		return 1, 1
	}
	if sum%limit == 0 {
		maxPage = sum / limit
	} else {
		maxPage = sum/limit + 1
	}
	return minPage, maxPage
}
