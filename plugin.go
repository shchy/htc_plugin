package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

type driveStatus int

const (
	nml driveStatus = iota
	war
	cpy
	cpi
	rsv
	fai
	blk
	unknown
)

func (s driveStatus) toString() string {
	switch s {
	case nml:
		return "NML"
	case war:
		return "WAR"
	case cpy:
		return "CPY"
	case cpi:
		return "CPI"
	case rsv:
		return "RSV"
	case fai:
		return "FAI"
	case blk:
		return "BLK"
	default:
		return "Unknown"
	}
}

func fromString(v string) driveStatus {
	switch v {
	case "NML":
		return nml
	case "WAR":
		return war
	case "CPY":
		return cpy
	case "CPI":
		return cpi
	case "RSV":
		return rsv
	case "FAI":
		return fai
	case "BLK":
		return blk
	default:
		return unknown
	}
}

type driveInfo struct {
	serialNumber           string
	status                 driveStatus
	usedEnduranceIndicator int
}

var graphdef = map[string]mp.Graphs{
	"hitachi.drive.status.#": {
		Label: "Drive Status",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: nml.toString(), Label: "Normal"},
			{Name: war.toString(), Label: "Warning"},
			{Name: cpy.toString(), Label: "Copy"},
			{Name: cpi.toString(), Label: "CopyIncomplete"},
			{Name: rsv.toString(), Label: "RSV"},
			{Name: fai.toString(), Label: "Fail"},
			{Name: blk.toString(), Label: "BLK"},
			{Name: unknown.toString(), Label: "Unknown"},
		},
	},
	"hitachi.drive.used.#": {
		Label: "Drive used Endurance Indicator(%)",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "used", Label: "Used"},
		},
	},
}

// Plugin プラグインの型
type Plugin struct {
	Host     string
	UserID   string
	Password string
}

func getDrives(host string, id string, pass string) ([]driveInfo, error) {
	url := fmt.Sprintf("http://%s/ConfigurationManager/v1/objects/drives?detailInfoType=usedEnduranceIndicator", host)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(id, pass)

	// リクエストの送信
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	drives := data["data"].([]interface{})
	infos := make([]driveInfo, len(drives))
	for i, drive := range drives {
		driveMap := drive.(map[string]interface{})

		status := driveMap["status"].(string)
		id := driveMap["serialNumber"].(string)
		ind := 0

		if v, ok := driveMap["usedEnduranceIndicator"]; ok {
			ind = int(v.(float64))
		}

		infos[i] = driveInfo{
			usedEnduranceIndicator: ind,
			serialNumber:           id,
			status:                 fromString(status),
		}
	}
	return infos, nil
}

// GraphDefinition グラフ定義
func (p Plugin) GraphDefinition() map[string]mp.Graphs {
	return graphdef
}

func getMetric(graphKey string, metric mp.Metrics, info driveInfo) float64 {
	if graphKey == "hitachi.drive.status.#" {
		if metric.Name == info.status.toString() {
			return 1
		}
		return 0
	} else if graphKey == "hitachi.drive.used.#" {
		return float64(info.usedEnduranceIndicator)
	}
	return 0
}

// FetchMetrics metricsの取得
func (p Plugin) FetchMetrics() (map[string]float64, error) {
	stat := make(map[string]float64)

	infos, err := getDrives(p.Host, p.UserID, p.Password)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		for graphKey := range graphdef {
			for _, metric := range graphdef[graphKey].Metrics {
				metricKey := strings.Replace(graphKey+"."+metric.Name, "#", info.serialNumber, -1)
				value := getMetric(graphKey, metric, info)
				stat[metricKey] = value
			}
		}
	}

	return stat, nil
}

func (p Plugin) do() {

	optHost := flag.String("host", "", "Api Host")
	optUserID := flag.String("userid", "", "User ID")
	optPassword := flag.String("password", "", "Password")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	p.Host = *optHost
	p.UserID = *optUserID
	p.Password = *optPassword

	helper := mp.NewMackerelPlugin(p)
	helper.Tempfile = *optTempfile

	helper.Run()
}
