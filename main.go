package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"sync"
	"time"
)

func main() {
	var (
		ctx          = context.Background()
		brokerVar, _ = g.Cfg().Get(ctx, "broker")
		broker       = brokerVar.Strings()
		brokerNum    = len(broker)

		targetVar, _ = g.Cfg().Get(ctx, "target")
		target       = targetVar.String()
		numVar, _    = g.Cfg().Get(ctx, "num")
		num          = numVar.Int()
		log          = glog.New()
		wg           = sync.WaitGroup{}
		res          = make(map[string]string, brokerNum)
		resLock      sync.RWMutex
	)

	fmt.Printf("开始时间: %s\n", time.Now().Format(time.DateTime))

	log.SetConfigWithMap(g.Map{
		"path":   "./log-main",
		"stdout": false,
	})

	wg.Add(brokerNum)
	for i := 0; i < brokerNum; i++ {
		go func(i int) {
			defer wg.Done()
			rep, err := g.Client().Get(ctx, broker[i], fmt.Sprintf("target=%s&num=%d", target, num))
			if err != nil {
				log.Error(ctx, err)
				return
			}
			defer rep.Close()
			resLock.Lock()
			res[broker[i]] = rep.ReadAllString()
			resLock.Unlock()
		}(i)
	}

	wg.Wait()

	var (
		succSlice, failSlice, htmlSlice []int64
		resSlice                        = make(map[string][]int64)
	)

	for _, v := range res {
		json.Unmarshal([]byte(v), &resSlice)
		succSlice = append(succSlice, resSlice["succSlice"]...)
		failSlice = append(failSlice, resSlice["failSlice"]...)
		htmlSlice = append(htmlSlice, resSlice["htmlSlice"]...)
	}

	sort(succSlice)

	var (
		succSliceLen                                            = len(succSlice)
		succSliceLen99                                          = succSliceLen * 99 / 100
		succSliceLen95                                          = succSliceLen * 95 / 100
		succSliceLen90                                          = succSliceLen * 90 / 100
		succSliceLen70                                          = succSliceLen * 70 / 100
		succSliceLen50                                          = succSliceLen * 50 / 100
		succSliceLen30                                          = succSliceLen * 30 / 100
		succSliceLen10                                          = succSliceLen * 10 / 100
		t100, t99, t95, t90, t70, t50, t30, t10                 int64
		req100, req99, req95, req90, req70, req50, req30, req10 int64
		minReq, maxReq                                          int64
	)

	for i := 0; i < len(succSlice); i++ {
		t100 += succSlice[i]
		if i < succSliceLen99 {
			t99 += succSlice[i]
		}
		if i < succSliceLen95 {
			t95 += succSlice[i]
		}
		if i < succSliceLen90 {
			t90 += succSlice[i]
		}
		if i < succSliceLen70 {
			t70 += succSlice[i]
		}
		if i < succSliceLen50 {
			t50 += succSlice[i]
		}
		if i < succSliceLen30 {
			t30 += succSlice[i]
		}
		if i < succSliceLen10 {
			t10 += succSlice[i]
		}
	}

	if succSliceLen > 0 {
		req100 = t100 / int64(succSliceLen)
		req99 = t99 / int64(succSliceLen99)
		req95 = t95 / int64(succSliceLen95)
		req90 = t90 / int64(succSliceLen90)
		req70 = t70 / int64(succSliceLen70)
		req50 = t50 / int64(succSliceLen50)
		req30 = t30 / int64(succSliceLen30)
		req10 = t10 / int64(succSliceLen50)
		minReq = succSlice[0]
		maxReq = succSlice[len(succSlice)-1]
	}

	fmt.Printf("样本: %v\n", num*brokerNum)
	fmt.Printf("success: %v\n", succSliceLen)
	fmt.Printf("fail: %v\n", fmt.Sprintf("%d(包含html错误%d个)", len(failSlice), len(htmlSlice)))
	fmt.Printf("平均响应时间: %v\n", req100)
	fmt.Printf("99%%响应: %v\n", req99)
	fmt.Printf("95%%响应: %v\n", req95)
	fmt.Printf("90%%响应: %v\n", req90)
	fmt.Printf("70%%响应: %v\n", req70)
	fmt.Printf("50%%响应: %v\n", req50)
	fmt.Printf("30%%响应: %v\n", req30)
	fmt.Printf("10%%响应: %v\n", req10)
	fmt.Printf("最小响应时间: %v\n", minReq)
	fmt.Printf("最大响应时间: %v\n", maxReq)
	fmt.Printf("请求目标: %v\n", target)
}

func sort(slice []int64) {
	count := len(slice)
	for i := 0; i < count; i++ {
		for j := 0; j < count-i-1; j++ {
			if slice[j] > slice[j+1] {
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
	}
}
