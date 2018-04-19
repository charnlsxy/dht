package main

import (
"encoding/hex"
"encoding/json"
"fmt"
_ "net/http/pprof"
	"me/study/btsearch/logft"
	"me/opensource/dht"
	"time"
)

type file struct {
	Path   []interface{} `json:"path"`
	Length int           `json:"length"`
}

type bitTorrent struct {
	InfoHash string `json:"infohash"`
	Name     string `json:"name"`
	Files    []file `json:"files,omitempty"`
	Length   int    `json:"length,omitempty"`
}

func main() {
	w := dht.NewWire(65536, 1024, 256)
	go func() {
		for resp := range w.Response() {
			metadata, err := dht.Decode(resp.MetadataInfo)
			if err != nil {
				logft.Error(err.Error())
				continue
			}
			info := metadata.(map[string]interface{})

			if _, ok := info["name"]; !ok {
				continue
			}

			bt := bitTorrent{
				InfoHash: hex.EncodeToString(resp.InfoHash),
				Name:     info["name"].(string),
			}

			if v, ok := info["files"]; ok {
				files := v.([]interface{})
				bt.Files = make([]file, len(files))

				for i, item := range files {
					f := item.(map[string]interface{})
					bt.Files[i] = file{
						Path:   f["path"].([]interface{}),
						Length: f["length"].(int),
					}
				}
			} else if _, ok := info["length"]; ok {
				bt.Length = info["length"].(int)
			}

			data, err := json.Marshal(bt)
			if err == nil {
				fmt.Printf("%s\n\n", data)
			}else{
				logft.Error(err.Error())
			}
		}
	}()
	go w.Run()

	config := dht.NewCrawlConfig()
	config.Address = ":9452"
	d := dht.New(config)

	go d.Run()
	//go w.Request([]byte("a95389905bda001d2dc688a1f71e18e736fe5efc"), ip, port)


	for {
		peers, err := d.GetPeers("a95389905bda001d2dc688a1f71e18e736fe5efc")
		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}
		fmt.Println("Found peers:", peers[0].IP.String())
		for _,p := range peers{
			w.Request([]byte("a95389905bda001d2dc688a1f71e18e736fe5efc"), p.IP.String(), p.Port)
		}
	}

}

