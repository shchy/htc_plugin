package main

import (
	"flag"
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
		return "nml"
	case war:
		return "war"
	case cpy:
		return "cpy"
	case cpi:
		return "cpi"
	case rsv:
		return "rsv"
	case fai:
		return "fai"
	case blk:
		return "blk"
	case unknown:
		fallthrough
	default:
		return "unknown"
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

func getDrives() ([]driveInfo, error) {
	// TODO HTTP
	infos := []driveInfo{
		driveInfo{
			serialNumber:           "asdf",
			status:                 nml,
			usedEnduranceIndicator: 0,
		},
		driveInfo{
			serialNumber:           "qwer",
			status:                 fai,
			usedEnduranceIndicator: 10,
		},
		driveInfo{
			serialNumber:           "zxcv",
			status:                 cpi,
			usedEnduranceIndicator: 99,
		},
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

	infos, err := getDrives()
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

	// optRegion := flag.String("region", "", "AWS Region")
	// optAccessKeyID := flag.String("access-key-id", "", "AWS Access Key ID")
	// optSecretAccessKey := flag.String("secret-access-key", "", "AWS Secret Access Key")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	helper := mp.NewMackerelPlugin(p)
	helper.Tempfile = *optTempfile

	helper.Run()
}
