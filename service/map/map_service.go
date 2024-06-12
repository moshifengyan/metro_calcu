package map_service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"metro_calcu/service/cache"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"metro_calcu/util"

	"github.com/samber/lo"
)

type RouteResp struct {
	Status  int64        `json:"status"`
	Message string       `json:"message"`
	Result  *RouteDetail `json:"result"`
}

type Step struct {
	Vehicle struct {
		EndName string `json:"end_name"`
	} `json:"vehicle"`
}

type Route struct {
	Duration int64     `json:"duration"`
	Steps    [][]*Step `json:"steps"`
}

type RouteDetail struct {
	Dest struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"destination"`
	Origin struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"origin"`
	Routes []*Route `json:"routes"`
}

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

type CacheSearchDest struct {
	EndName  string
	Duration int64
}

func GetStationsByPage(page int64, name, region string, isLine bool) ([]*StationsDetail, error) {
	// 设置 API 参数
	path := "/place/v2/search"
	params := [][]string{
		{"query", name},
		{"tag", "地铁站,地铁/轻轨,地铁\\轻轨"},
		{"region", region}, // 深圳
		{"output", "json"},
		{"ak", util.AK},
		{"page_size", "20"},
		{"page_num", strconv.Itoa(int(page))},
	}
	paramsStr, sn := util.CalculateSN(params, path, util.SK)

	// 发起请求
	request, err := url.Parse(util.MapHost + path + "?" + paramsStr + "&sn=" + sn)
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

func GetStartLocation(region, name string) (float64, float64, string) {
	stations, err := GetStationsByPage(0, name, region, false)
	if err != nil || len(stations) == 0 {
		return 0.0, 0.0, ""
	}
	station := stations[0]
	loc := station.Location
	return loc.Lat, loc.Lng, station.UID

}

func SearchDestStation(startLat, startLng, destLat, destLng float64, timeDur int64, startUID, destUID string) (string, int64) {
	cacheKey := startUID + destUID
	cacheVal := cache.Get(cacheKey)
	if cacheVal != nil {
		cv := cacheVal.(CacheSearchDest)
		return cv.EndName, cv.Duration
	}
	path := "/directionlite/v1/transit"

	// 设置请求参数
	params := [][]string{
		{"origin", fmt.Sprintf("%v,%v", startLat, startLng)},
		{"destination", fmt.Sprintf("%v,%v", destLat, destLng)},
		{"ak", util.AK},
		{"timestamp", fmt.Sprintf("%d", time.Now().Unix())},
	}
	if startUID != "" && destUID != "" {
		params = append(params, []string{"origin_uid", startUID}, []string{"destination_uid", destUID})
	}
	paramsStr, sn := util.CalculateSN(params, path, util.SK)

	// 发起请求
	request, err := url.Parse(util.MapHost + path + "?" + paramsStr + "&sn=" + sn)
	if nil != err {
		fmt.Printf("host error: %v", err)
		return "", 0
	}

	resp, err1 := http.Get(request.String())
	fmt.Printf("url: %s\n", request.String())
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("request error: %v", err1)
		return "", 0
	}
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("response error: %v", err2)
	}
	var routeResp RouteResp
	err = json.Unmarshal(body, &routeResp)
	if err != nil {
		return "", 0
	}
	if routeResp.Result == nil || len(routeResp.Result.Routes) == 0 {
		return "", 0
	}
	routes := routeResp.Result.Routes
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Duration < routes[j].Duration
	})
	selectedRoute := routes[0]
	if selectedRoute.Duration > timeDur {
		return "", 0
	}
	curRouteSteps := selectedRoute.Steps
	for i := len(curRouteSteps) - 1; i >= 0; i-- {
		if strings.Contains(curRouteSteps[i][0].Vehicle.EndName, "站") {

			return curRouteSteps[i][0].Vehicle.EndName, selectedRoute.Duration
		}
	}
	return "", 0
}
