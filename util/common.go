package util

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

const MapHost = "https://api.map.baidu.com"
const AK = "vx8BaTvDoFdY2x1grejbHR1FOoznyTSP"
const SK = "zKZ16sLgw6vKerUmDr31bUKHjhcAAM7O"
const CacheKey = "all_station"

var MetroLines = map[string][]string{
	"深圳": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "16", "20"},
	"广州": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "13", "14", "18", "21", "22"},
}

type DetRet struct {
	Name     string
	Duration int64
}

// SliceSafer 是一个线程安全的 slice 包装器
type SliceSafer struct {
	slice []DetRet
	mutex sync.Mutex
}

// NewSliceSafer 创建一个新的线程安全 slice 实例
func NewSliceSafer() *SliceSafer {
	return &SliceSafer{
		slice: make([]DetRet, 0),
	}
}

// Append 线程安全地向 slice 追加元素
func (s *SliceSafer) Append(values ...DetRet) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.slice = append(s.slice, values...)
}

func (s *SliceSafer) GetSlice() []DetRet {
	return s.slice
}

func FormatTime(dur int64) string {
	duration := time.Duration(dur) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	remainingSeconds := int(duration.Seconds()) % 60

	if hours == 0 {
		return fmt.Sprintf("%d分%d秒", minutes, remainingSeconds)
	}
	return fmt.Sprintf("%d小时%d分%d秒", hours, minutes, remainingSeconds)
}

func CalculateSN(params [][]string, path, sk string) (string, string) {
	paramsArr := make([]string, 0)
	for _, v := range params {
		kv := v[0] + "=" + (v[1])
		paramsArr = append(paramsArr, kv)
	}
	paramsStr := strings.Join(paramsArr, "&")

	// 计算sn
	queryStr := url.QueryEscape(path + "?" + paramsStr)
	str := queryStr + sk
	key := md5.Sum([]byte(str))
	sn := fmt.Sprintf("%x", key)
	return paramsStr, sn
}
