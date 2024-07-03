package Captcha

import (
	"BoostTool/Core/Utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func Capmonster(apikey, host, auth, webkey, rqdata, token1 string) string {
	client := http.Client{Timeout: time.Second * 30}
	var CapmonsterTask capmonsterTaskID
	payload := map[string]interface{}{
		"clientKey": apikey,
		"task": map[string]interface{}{
			"type":          "HCaptchaTask",
			"websiteURL":    "https://discord.com/",
			"websiteKey":    webkey,
			"userAgent":     "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9017 Chrome/108.0.5359.215 Electron/22.3.12 Safari/537.36",
			"proxyType":     "http",
			"proxyAddress":  strings.Split(host, ":")[0],
			"proxyPort":     strings.Split(host, ":")[1],
			"proxyLogin":    strings.Split(auth, ":")[0],
			"proxyPassword": strings.Split(auth, ":")[1],
			"data":          rqdata,
		},
	}

	jsonpay, _ := json.Marshal(payload)

	req1, _ := http.NewRequest("POST", "https://api.capmonster.cloud/createTask", bytes.NewBuffer(jsonpay))
	req1.Header.Set("Content-Type", "application/json")
	resp1, _ := client.Do(req1)
	body, _ := io.ReadAll(resp1.Body)
	_ = json.Unmarshal(body, &CapmonsterTask)

	if CapmonsterTask.ErrorID != 1 {
		Utils.LogInfo("Captcha Task", "Task ID", fmt.Sprintf("%v", CapmonsterTask.TaskID))
		for i := 0; i < 10; i++ {
			var CapmonsterResponse capmonsterGetTask
			payload2 := map[string]interface{}{
				"clientKey": apikey,
				"taskId":    CapmonsterTask.TaskID,
			}
			jsonpayload, _ := json.Marshal(payload2)
			req2, _ := http.NewRequest("POST", "https://api.capmonster.cloud/getTaskResult", bytes.NewBuffer(jsonpayload))
			req2.Header.Set("Content-Type", "application/json")
			resp2, _ := client.Do(req2)
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)

			_ = json.Unmarshal(body2, &CapmonsterResponse)

			if CapmonsterResponse.Status != "ready" && CapmonsterResponse.Status != "processing" {
				Utils.LogError("Couldn't Solved Captcha, Retrying", "Token", token1)
				return ""
			} else if CapmonsterResponse.Status == "ready" {
				return CapmonsterResponse.Solution.GRecaptchaResponse
			} else {
				time.Sleep(time.Second * 2)
				continue
			}
		}
	} else {
		Utils.LogError("Couldn't Get Captcha Task ID, Check API Key or Contact Support", "Error", string(body))
	}

	return ""
}
