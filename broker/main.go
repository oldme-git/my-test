package main

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
	"time"
)

func main() {
	s := g.Server()
	s.BindHandler("/", func(r *ghttp.Request) {
		var (
			ctx    = context.Background()
			target = r.Get("target").String()
			num    = r.Get("num").Int()
			log    = glog.New()
			wg     = sync.WaitGroup{}

			succLock, failLock, htmlLock    sync.RWMutex
			succSlice, failSlice, htmlSlice []int64
		)
		//ctx, _ = context.WithTimeout(ctx, 10*time.Second)

		log.SetConfigWithMap(g.Map{
			"path":   "./log",
			"stdout": false,
		})

		wg.Add(num)
		for i := 0; i < num; i++ {
			go func() {
				defer wg.Done()
				var start time.Time
				trace := &httptrace.ClientTrace{
					WroteRequest: func(info httptrace.WroteRequestInfo) { start = time.Now() },
				}
				rep, err := g.Client().Use(func(c *gclient.Client, r *http.Request) (*gclient.Response, error) {
					r = r.WithContext(httptrace.WithClientTrace(r.Context(), trace))
					return c.Next(r)
				}).Get(ctx, target)
				elapsed := time.Since(start).Milliseconds()
				if err != nil {
					log.Info(ctx, err.Error())
					failLock.Lock()
					failSlice = append(failSlice, elapsed)
					failLock.Unlock()
					return
				}
				defer rep.Close()
				if rep.StatusCode != http.StatusOK {
					log.Info(ctx, fmt.Sprintf("非200端口%d", rep.StatusCode))
					failLock.Lock()
					failSlice = append(failSlice, elapsed)
					failLock.Unlock()
					return
				}
				res := rep.ReadAllString()
				if strings.Contains(res, "<!DOCTYPE html>") {
					log.Info(ctx, res)
					htmlLock.Lock()
					htmlSlice = append(htmlSlice, elapsed)
					htmlLock.Unlock()
					return
				}
				succLock.Lock()
				succSlice = append(succSlice, elapsed)
				succLock.Unlock()
			}()
		}

		wg.Wait()

		failSlice = append(failSlice, htmlSlice...)

		r.Response.WriteJson(map[string][]int64{
			"succSlice": succSlice,
			"failSlice": failSlice,
			"htmlSlice": htmlSlice,
		})
	})
	s.Run()
}
