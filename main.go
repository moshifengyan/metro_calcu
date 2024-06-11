package main

import (
	"fmt"
	"sort"
	"sync"
)

type CalcuRet struct {
	Name       string
	FormatTime string
}

type TransitStationsResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Stations []string `json:"stations"`
	} `json:"result"`
}

func main() {
	res := NewSliceSafer()
	allStations := getAllStations("shenzhen")
	//fmt.Printf("%v", lo.Map(allStations, func(item *StationsDetail, index int) string {
	//	return item.Name + "\n"
	//}))
	startLat, startLng, startUID := getStartLocation("高新园")
	wg := sync.WaitGroup{}
	limit := make(chan bool, 5)
	for _, station := range allStations {
		wg.Add(1)
		limit <- true
		go func(startUID string, station *StationsDetail) {
			defer func() {
				wg.Done()
				<-limit
			}()
			destLat, destLng, destUID := station.Location.Lat, station.Location.Lng, station.UID
			det, dur := searchDestStation(startLat, startLng, destLat, destLng, 3600, startUID, destUID)
			if det == "" {
				return
			}
			res.Append(DetRet{
				det,
				dur,
			})
		}(startUID, station)

	}
	wg.Wait()
	sort.Slice(res.slice, func(i, j int) bool {
		return res.slice[i].Duration > res.slice[j].Duration
	})
	for _, ret := range res.slice {
		fmt.Printf("到%v %v\n", ret.Name, FormatTime(ret.Duration))
	}
}
