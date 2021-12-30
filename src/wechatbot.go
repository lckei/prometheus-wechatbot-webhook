package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"text/template"
	"time"

	"github.com/golang/glog"
)

type KV map[string]string
type Data struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   Alerts `json:"alerts"`

	GroupLabels       KV `json:"groupLabels"`
	CommonLabels      KV `json:"commonLabels"`
	CommonAnnotations KV `json:"commonAnnotations"`

	ExternalURL string `json:"externalURL"`
}

// Alert holds one alert for notification templates.
type Alert struct {
	Status       string    `json:"status"`
	Labels       KV        `json:"labels"`
	Annotations  KV        `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint"`
}

// Alerts is a list of Alert objects.
type Alerts []Alert

func main() {
	http.HandleFunc("/wechatbot", func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var t Data

		err := decoder.Decode(&t)
		if err != nil {
			glog.Error(err)
		}

		tmpl := template.Must(template.ParseFiles("wechatbot.tmpl"))

		//@parma: tpl 获取模板字节内容
		var tpl bytes.Buffer
		if err := tmpl.Execute(&tpl, t); err != nil {
			glog.Error(err)
		}

		//@parma: wechatbotUrlBytes wechatbot api url with key
		var wechatbotUrlBytes bytes.Buffer
		if err := tmpl.ExecuteTemplate(&wechatbotUrlBytes, "wechatbot.url.api", "no data needed"); err != nil {
			glog.Error(err)
			return
		}

		//post the context to the wechat api
		postBody, _ := json.Marshal(map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]interface{}{
				"content": tpl.String(),
			},
		})
		responseBody := bytes.NewBuffer(postBody)
		resp, err := http.Post(wechatbotUrlBytes.String(), "application/json", responseBody)
		if err != nil {
			glog.Error(err)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			}
			glog.Error("Broken : ", string(body))
			w.WriteHeader(http.StatusBadRequest)
			w.Write(body)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("to wechatbot success!"))
		}

	})

	http.ListenAndServe(":9080", nil)
}
