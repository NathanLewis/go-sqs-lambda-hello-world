package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport"
	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/util"
	//"github.com/prometheus/client_golang/prometheus/promhttp"
)

var config = util.Configuration{}

func main() {

	configFile, _ := os.Open("conf.json")
	decoder := json.NewDecoder(configFile)
	err := decoder.Decode(&config)
	if err != nil {
		panic("Error reading conf")
	}
	//	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/zip", handler)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	// Get "ref" URL params
	campaignIds, ok := r.URL.Query()["campaignId"]
	if !ok || len(campaignIds) < 1 {
		http.Error(w, "S3 File Zipper. Pass ?ref= to use.", 500)
		return
	}
	w.Header().Add("Content-Disposition", "attachment; filename=\"download.zip\"")
	w.Header().Add("Content-Type", "application/zip")

	campaignID := campaignIds[0]
	s3Processor := zipexport.ZipProcessor{}
	s3Processor.Process(w, campaignID)
	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}
