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
	const lic = "LTZQOSJHPE-eyJhc3NpZ25lZUVtYWlsIjoiIiwiYXNzaWduZWVOYW1lIjoiIiwiYXV0b1Byb2xvbmdhdGVkIjpmYWxzZSwiY2hlY2tDb25jdXJyZW50VXNlIjpmYWxzZSwiZ3JhY2VQZXJpb2REYXlzIjo3LCJoYXNoIjoiVFJJQUw6LTE2MzUyMTY1NzgiLCJsaWNlbnNlSWQiOiJMVFpRT1NKSFBFIiwibGljZW5zZVJlc3RyaWN0aW9uIjoiIiwibGljZW5zZWVOYW1lIjoiamluYmFvemkiLCJtZXRhZGF0YSI6IjAxMjAyMzA5MTRQU0FYMDAwMDA1IiwicHJvZHVjdHMiOlt7ImNvZGUiOiJJSSIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IkNMIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJjb2RlIjoiUFMiLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJHTyIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IlBDIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJjb2RlIjoiV1MiLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJSRCIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IkRCIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In0seyJjb2RlIjoiUk0iLCJleHRlbmQiOnRydWUsImZhbGxiYWNrRGF0ZSI6IjIwOTktMDktMTQiLCJwYWlkVXBUbyI6IjIwOTktMDktMTQifSx7ImNvZGUiOiJBQyIsImV4dGVuZCI6dHJ1ZSwiZmFsbGJhY2tEYXRlIjoiMjA5OS0wOS0xNCIsInBhaWRVcFRvIjoiMjA5OS0wOS0xNCJ9LHsiY29kZSI6IkRTIiwiZXh0ZW5kIjp0cnVlLCJmYWxsYmFja0RhdGUiOiIyMDk5LTA5LTE0IiwicGFpZFVwVG8iOiIyMDk5LTA5LTE0In1dfQ==-NZMh1r+9KcbFsxQ45Eb3NAwvKCu6DL3feYIxWC9q5rI/SbFHzlZq0ZU/oBiLkQGxYwfmui3MfzM/SvlcpNV3bOlOD72sDXifG8I9r9FvlH69kOZFrzKrmkppwgax6a9USBtZKjpHhkSGY0U2+KJqBr+zLLZ4JdX0Q6KmCQmduy7aFh8dZXCzsA17opO+eIItZby1SlxHNARjVOzToNv79r1b1ocmwUsZ94brGptO/uMcOzPziwQXV9bGiF4rOf9KbYr7/C5dwNdHTWeX3zM1UD0kVwV6NJ1t3gN/mIF/rN0/wziZfAdhJEvgfWby9hdz1Po5ghASlSMf8Qvf1Us2uOUS9kN4NR6sIVwMrIjrlhnbTah/f+Qc2TGTpQQP6VUBpQmQ2AonmYnxtdao+muTiGV0f7QZednP5HrDAn0RN+bq6pCN7A5n6ojT5CAiaH32CiSOboH9SxxT55shnlsBkMl4u6+zwTxwXp2LrvyUfwEWJmVTucMs89P7G62JamxnUe8V+1bGbNxqbOb7svWRnO4o3hYBQM2gpGaTQDNXccL44J10zpNZK4p9UM+yTKexgoR7feSqI2UYJFqsnl8qz/BT0Edhdiezq11lpI3mbz+TK2c34yLK4rdMYCebVsN2hacZbZ5P0tMic9EwAtizwnzyoBUn2exs9zGck14I/QM=-MIIEpzCCAo+gAwIBAgIEZ3dzozANBgkqhkiG9w0BAQsFADAYMRYwFAYDVQQDDA1KZXRQcm9maWxlIENBMB4XDTI0MDEwMzA1MjAzNVoXDTM1MDEwMzA1MjAzNVowEzERMA8GA1UEAwwISmluYmFvemkwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCG/J39JeLwyMfR73wIHm4gXoW3RkcJd7Pj8xQiCwUz5Hj+wBmBy1SOGP2wKoZFnZEx4AgTqJWEPKNwZNjuukCJkuBNvucBhJ2qsg4om/up19NRkE/bdvrm56T1eMXmEwV8EHtDpAuDAyJYTfHIkr87tM7bcgookyv5MdSebK/M1BPE1lTeEH3/9IJXem3zWOB0eMw88P7G5z1NLecSpiAAWmHU7MhB6KU27l5T1O0uKOKJGEnvD6EeP1ZwCD/s0YqNHSJzekYdoGnt1HJ5JTz4JqjayNuYgrkgaYIUBLuEfEyKIvwzvE+BauSgyjudMaOGuLZRhIqWLAjt0gMKOLO/pRD5HyOrZWktbZ/Vi0YqMbm4fdinMzPABGIewsgguKfcvkfAf0Ni6t4uPFpFH49XUDRvjQfL+p+RL7M3lQ1VpGS9UCqM+/kYoUeJDRP+1kErXMbo19ymYvfTVvMchBl8OXwqORjolNfCtTJkDbf7eWMwzxsDznl6S5xbvn1JLOWfnIf88gnH2YIh9SzW2N/BIy0C4j1bMzxkjM6/85pbqL+5LMvLYom5z6E3hf5MJdJWCogBVUNrUkUk0h3863iRR+ILvxBxrxgiOoJENHuU1zJa0n02PO4wCek+E/lskFfUftRp4rE8GQ6iJmpDkER918zPihj7C7JJjXKT6Ou8+QIDAQABMA0GCSqGSIb3DQEBCwUAA4ICAQACk4B/bYgljpVi2w1u2V+8pGUkn9VxfK04owmGWKwdJ+twUzF1tAzcggJ+D4/FlzmVLDgR+V/nEeVLuRKbZ0pVnfmFEQY5FFkRUmD/+EHjM8BbooOQ3FFHeZjDBOAILNY5ZIVggvcq9vDiNqTBMHmnzIYtY6GJ6bLkX6mvIv/xprldRmjydKKpkwC4aeHrMx1x0KHu0rTAH7onjWusnARERc+ruNTZje4JZ6OTLvOUP0F/FMvuXinMHb1Q1u6XgGfT4cXAK85uf9gjc+ck7il0cS5JCecKAHo1OZlnzhd/D0HEhNwxlRjofccLPgebRoiI0TM+aYQjnOF3z6EAsKjxKWA0IGJ7H1mO1GpfcYxRvDQp64Au2H4hMjjN6qZt6Y1BAJqu5g+36MiCYAOUn2F4mayOLG/90XDEvfvCGwuK0J5YsxZ0rDvf2K98qkRgGthI1LuUltdyqS0d3JTLdvA5u1xPiifCtg/WbX7Dl5KnLi9jdO50MLNDLd/gIPaVFFWvAOBjl0dZKH582Sy9JWtHPUaI9NrgMzq9e50BchYwdPKYBB07tCL1VKI1FWKSuOnA+oEUgYVxwIwQNCxaOwHX9RUfjeY2ZEJJvzJS4lP9F1Ps3LvQydFx6ARcXTKoQgViCyiTV56LvRlLcXlyhcIf3iTn7zzqifvrR9yqKOVXMw=="
	return true, lic
}
