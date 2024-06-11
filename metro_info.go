package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type StationsDetail struct {
	Name     string `json:"name"`
	UID      string `json:"uid"`
	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
	Address    string `json:"address"`
	Telephone  string `json:"telephone"`
	DetailInfo struct {
		Tag string `json:"tag"`
	} `json:"detail_info"`
}

type PlaceSearchResponse struct {
	Status  int               `json:"status"`
	Message string            `json:"message"`
	Results []*StationsDetail `json:"results"`
}

func getAllStations(region string) []*StationsDetail {
	allStations := make([]*StationsDetail, 0)
	lineNums, ok := MetroLines[region]
	if !ok {
		return nil
	}
	for _, num := range lineNums {
		var page int64
		lineNumName := "地铁" + num + "号线"
		for {
			stations, err := getStationsByPage(page, lineNumName, true)
			if err != nil {
				panic(err)
			}
			if len(stations) == 0 {
				break
			}
			page += 1
			allStations = append(allStations, stations...)
		}
	}
	allStations = lo.Filter(allStations, func(item *StationsDetail, index int) bool {
		return item.Address != ""
	})
	allStationsMap := lo.SliceToMap(allStations, func(item *StationsDetail) (string, *StationsDetail) {
		return item.UID, item
	})
	return lo.Values(allStationsMap)
}

func getStationsByPage(page int64, name string, isLine bool) ([]*StationsDetail, error) {
	// 设置 API 参数
	path := "/place/v2/search"
	params := [][]string{
		{"query", name},
		{"tag", "地铁站,地铁/轻轨,地铁\\轻轨"},
		{"region", "340"}, // 深圳
		{"output", "json"},
		{"ak", AK},
		{"page_size", "20"},
		{"page_num", strconv.Itoa(int(page))},
	}
	paramsStr, sn := calculateSN(params, path, SK)

	// 发起请求
	request, err := url.Parse(MapHost + path + "?" + paramsStr + "&sn=" + sn)
	if nil != err {
		fmt.Printf("host error: %v", err)
		return nil, err
	}
	resp, err1 := http.Get(request.String())
	// 请注意，此处打印的url为非urlencode后的请求串
	// 如果将该请求串直接粘贴到浏览器中发起请求，由于浏览器会自动进行urlencode，会导致返回sn校验失败
	fmt.Printf("url: %s\n", request.String())
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("request error: %v", err1)
		return nil, err
	}
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("response error: %v", err2)
	}
	var stations PlaceSearchResponse
	err = json.Unmarshal(body, &stations)
	if err != nil {
		return nil, err
	}

	res := lo.Filter(stations.Results, func(item *StationsDetail, index int) bool {
		if isLine {
			addresses := strings.Split(item.Address, ",") //地铁1号线，地铁2号线
			return lo.Contains(addresses, name)
		}
		return item.Name == name
	})
	return res, nil
}

func calculateSN(params [][]string, path, sk string) (string, string) {
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
