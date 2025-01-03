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
	const lic = "AUCPLZYALL-eyJhc3NpZ25lZUVtYWlsIjoiIiwiYXNzaWduZWVOYW1lIjoiIiwiYXV0b1Byb2xvbmdhdGVkIjpmYWxzZSwiY2hlY2tDb25jdXJyZW50VXNlIjpmYWxzZSwiZ3JhY2VQZXJpb2REYXlzIjo3LCJoYXNoIjoiVFJJQUw6LTE2MzUyMTY1NzgiLCJsaWNlbnNlSWQiOiJBVUNQTFpZQUxMIiwibGljZW5zZVJlc3RyaWN0aW9uIjoiIiwibGljZW5zZWVOYW1lIjoibGxqIiwibWV0YWRhdGEiOiIwMTIwMjMwOTE0UFNBWDAwMDAwNSIsInByb2R1Y3RzIjpbeyJjb2RlIjoiSUkiLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJDTCIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IlBTIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJjb2RlIjoiR08iLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJQQyIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IldTIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJjb2RlIjoiUkQiLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJEQiIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6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\n"
	return true, lic
}
