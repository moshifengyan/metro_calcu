package manager

import (
	"fmt"
	"metro_calcu/service/cache"
	map_service "metro_calcu/service/map"
	"metro_calcu/util"
	"sort"
	"sync"

	"github.com/samber/lo"
)

func CalculateDest(region, startStation string, duration int64) []string {
	returnRes := make([]string, 0)
	res := util.NewSliceSafer()
	allStations := getAllStations(region)
	//fmt.Printf("%v", lo.Map(allStations, func(item *StationsDetail, index int) string {
	//	return item.Name + "\n"
	//}))
	startLat, startLng, startUID := map_service.GetStartLocation(region, startStation)
	wg := sync.WaitGroup{}
	limit := make(chan bool, 5)
	for _, station := range allStations {
		wg.Add(1)
		limit <- true
		go func(startUID string, station *map_service.StationsDetail) {
			defer func() {
				wg.Done()
				<-limit
			}()
			destLat, destLng, destUID := station.Location.Lat, station.Location.Lng, station.UID
			det, dur := map_service.SearchDestStation(startLat, startLng, destLat, destLng, duration, startUID, destUID)
			if det == "" {
				return
			}
			res.Append(util.DetRet{
				det,
				dur,
			})
		}(startUID, station)

	}
	wg.Wait()
	sort.Slice(res.GetSlice(), func(i, j int) bool {
		return res.GetSlice()[i].Duration > res.GetSlice()[j].Duration
	})
	for _, ret := range res.GetSlice() {
		returnRes = append(returnRes, fmt.Sprintf("到%v %v\n", ret.Name, util.FormatTime(ret.Duration)))
	}
	return returnRes
}

func getAllStations(region string) []*map_service.StationsDetail {
	cacheVal := cache.Get(util.CacheKey)
	if cacheVal != nil {
		return cacheVal.([]*map_service.StationsDetail)
	}
	allStations := make([]*map_service.StationsDetail, 0)
	lineNums, ok := util.MetroLines[region]
	if !ok {
		return nil
	}
	for _, num := range lineNums {
		var page int64
		lineNumName := "地铁" + num + "号线"
		for {
			stations, err := map_service.GetStationsByPage(page, lineNumName, region, true)
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
	allStations = lo.Filter(allStations, func(item *map_service.StationsDetail, index int) bool {
		return item.Address != ""
	})
	allStationsMap := lo.SliceToMap(allStations, func(item *map_service.StationsDetail) (string, *map_service.StationsDetail) {
		return item.UID, item
	})
	retAll := lo.Values(allStationsMap)
	cache.Set(util.CacheKey, retAll)
	return lo.Values(allStationsMap)
}
