package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"time"
)

type Client struct {
	Hosts []string // 服务器地址s
	host  string   // 检查后的服务器地址
}

func (c *Client) SetProxy(lang string) {
	defer c.setHost()
	proxy := httplib.BeegoHTTPSettings{}.Proxy
	proxyText := ""
	if os.Getenv("http_proxy") != "" {
		proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(os.Getenv("http_proxy"))
		}
		proxyText = os.Getenv("http_proxy") + " " + tr.Tr("经由") + " http_proxy " + tr.Tr("代理访问")
	}
	if os.Getenv("https_proxy") != "" {
		proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(os.Getenv("https_proxy"))
		}
		proxyText = os.Getenv("https_proxy") + " " + tr.Tr("经由") + " https_proxy " + tr.Tr("代理访问")
	}
	if os.Getenv("all_proxy") != "" {
		proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(os.Getenv("all_proxy"))
		}
		proxyText = os.Getenv("all_proxy") + " " + tr.Tr("经由") + " all_proxy " + tr.Tr("代理访问")
	}
	httplib.SetDefaultSetting(httplib.BeegoHTTPSettings{
		Proxy:            proxy,
		ReadWriteTimeout: 30 * time.Second,
		ConnectTimeout:   30 * time.Second,
		Gzip:             true,
		DumpBody:         true,
		UserAgent: fmt.Sprintf(`{"lang":"%s","GOOS":"%s","ARCH":"%s","version":%d,"deviceID":"%s","machineID":"%s"}`,
			lang, runtime.GOOS, runtime.GOARCH, version, deviceID, machineID),
	})
	if len(proxyText) > 0 {
		fmt.Printf(yellow, proxyText)
	}
}

func (c *Client) setHost() {
	c.host = c.Hosts[0]
	for _, v := range c.Hosts {
		_, err := httplib.Get(v).SetTimeout(4*time.Second, 4*time.Second).String()
		if err == nil {
			c.host = v
			return
		}
	}
	return
}

func (c *Client) GetAD() (ad string) {
	res, err := httplib.Get(c.host + "/ad").String()
	if err != nil {
		return
	}
	return res
}

func (c *Client) GetPayUrl() (payUrl, orderID string) {
	res, err := httplib.Get(c.host + "/payUrl").String()
	if err != nil {
		fmt.Println(err)
		return
	}
	payUrl = gjson.Get(res, "payUrl").String()
	orderID = gjson.Get(res, "orderID").String()
	return
}
func (c *Client) PayCheck(orderID, deviceID string) (isPay bool) {
	res, err := httplib.Get(c.host + "/payCheck?orderID=" + orderID + "&deviceID=" + deviceID).String()
	if err != nil {
		fmt.Println(err)
		return
	}
	isPay = gjson.Get(res, "isPay").Bool()
	return
}

func (c *Client) GetMyInfo(deviceID string) (sCount, sPayCount, isPay, ticket, exp string) {
	body, _ := json.Marshal(map[string]string{
		"device":    deviceID,
		"deviceMac": getMac_241018(),
		"sDevice":   getPromotion(),
	})
	dUser, _ := user.Current()
	deviceName := ""
	if dUser != nil {
		deviceName = dUser.Name
		if deviceName == "" {
			deviceName = dUser.Username
		}
	}
	res, err := httplib.Post(c.host+"/my").Body(body).Header("deviceName", deviceName).String()
	if err != nil {
		panic(fmt.Sprintf("\u001B[31m%s\u001B[0m", err))
		return
	}
	sCount = gjson.Get(res, "sCount").String()
	sPayCount = gjson.Get(res, "sPayCount").String()
	isPay = gjson.Get(res, "isPay").String()
	ticket = gjson.Get(res, "ticket").String()
	exp = gjson.Get(res, "exp").String()
	return
}

func (c *Client) CheckVersion(version string) (upUrl string) {
	res, err := httplib.Get(c.host + "/version?version=" + version + "&plat=" + runtime.GOOS + "_" + runtime.GOARCH).String()
	if err != nil {
		return ""
	}
	upUrl = gjson.Get(res, "url").String()
	return
}

func (c *Client) GetLic(product string, dur int) (isOk bool, result string) {
	const lic = "RXTKUYVQGQ-eyJhc3NpZ25lZUVtYWlsIjoiIiwiYXNzaWduZWVOYW1lIjoiIiwiYXV0b1Byb2xvbmdhdGVkIjpmYWxzZSwiY2hlY2tDb25jdXJyZW50VXNlIjpmYWxzZSwiZ3JhY2VQZXJpb2REYXlzIjo3LCJoYXNoIjoiVFJJQUw6LTE2MzUyMTY1NzgiLCJsaWNlbnNlSWQiOiJSWFRLVVlWUUdRIiwibGljZW5zZVJlc3RyaWN0aW9uIjoiIiwibGljZW5zZWVOYW1lIjoiamluYmFvemkiLCJtZXRhZGF0YSI6IjAxMjAyMzA5MTRQU0FYMDAwMDA1IiwicHJvZHVjdHMiOlt7ImNvZGUiOiJJSSIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IkNMIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJjb2RlIjoiUFMiLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJHTyIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IlBDIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJjb2RlIjoiV1MiLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJSRCIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IkRCIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJj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\n"
	return true, lic
}
