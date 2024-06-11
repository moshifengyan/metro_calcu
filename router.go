package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
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

func getStartLocation(name string) (float64, float64, string) {
	stations, err := getStationsByPage(0, name, false)
	if err != nil || len(stations) == 0 {
		return 0.0, 0.0, ""
	}
	station := stations[0]
	loc := station.Location
	return loc.Lat, loc.Lng, station.UID

}

func searchDestStation(startLat, startLng, destLat, destLng float64, timeDur int64, startUID, destUID string) (string, int64) {
	path := "/directionlite/v1/transit"

	// 设置请求参数
	params := [][]string{
		{"origin", fmt.Sprintf("%v,%v", startLat, startLng)},
		{"destination", fmt.Sprintf("%v,%v", destLat, destLng)},
		{"ak", AK},
		{"timestamp", fmt.Sprintf("%d", time.Now().Unix())},
	}
	if startUID != "" && destUID != "" {
		params = append(params, []string{"origin_uid", startUID}, []string{"destination_uid", destUID})
	}
	paramsStr, sn := calculateSN(params, path, SK)

	// 发起请求
	request, err := url.Parse(MapHost + path + "?" + paramsStr + "&sn=" + sn)
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
