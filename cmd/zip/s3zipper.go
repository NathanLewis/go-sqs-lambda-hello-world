package main

import (
	"archive/zip"
	"compress/flate"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cavaliercoder/grab"
	"golang.org/x/net/http2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	//"github.com/prometheus/client_golang/prometheus/promhttp"

	ugcAws "github.com/NathanLewis/go-sqs-lambda-hello-world/internal/app/zipexport/aws"
	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/app/zipexport/wormhole"
	ugchttp "github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/http"
	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/util"
)

//HTTPClientSettings snippet-start:[s3.go.customHttpClient_struct]
type HTTPClientSettings struct {
	Connect          time.Duration
	ConnKeepAlive    time.Duration
	ExpectContinue   time.Duration
	IdleConn         time.Duration
	MaxAllIdleConns  int
	MaxHostIdleConns int
	ResponseHeader   time.Duration
	TLSHandshake     time.Duration
}

// snippet-end:[s3.go.customHttpClient_struct]

//NewHTTPClientWithSettings snippet-start:[s3.go.customHttpClient_client]
func NewHTTPClientWithSettings(httpSettings HTTPClientSettings) *http.Client {
	tr := &http.Transport{
		ResponseHeaderTimeout: httpSettings.ResponseHeader,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: httpSettings.ConnKeepAlive,
			DualStack: true,
			Timeout:   httpSettings.Connect,
		}).DialContext,
		MaxIdleConns:          httpSettings.MaxAllIdleConns,
		IdleConnTimeout:       httpSettings.IdleConn,
		TLSHandshakeTimeout:   httpSettings.TLSHandshake,
		MaxIdleConnsPerHost:   httpSettings.MaxHostIdleConns,
		ExpectContinueTimeout: httpSettings.ExpectContinue,
	}

	// So client makes HTTP/2 requests
	http2.ConfigureTransport(tr)

	return &http.Client{
		Transport: tr,
	}
}

var config = util.Configuration{}
var downloader *s3manager.Downloader

func main() {

	configFile, _ := os.Open("conf.json")
	decoder := json.NewDecoder(configFile)
	err := decoder.Decode(&config)
	if err != nil {
		panic("Error reading conf")
	}

	initAwsBucket()

//	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/zip", handler)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}

func initAwsBucket() {

	downloader = awsDownloader()

}

//FakeWriterAt is Used to stream from s3
type FakeWriterAt struct {
	w io.Writer
}

//WriteAt is used for simulating s3 downloader calls
func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads
	return fw.w.Write(p)
}

// Remove all other unrecognised characters apart from
var makeSafeFileName = regexp.MustCompile(`[#<>:"/\|?*\\]`)

//S3ChannelData data used by the channel
type S3ChannelData struct {
	ItemToZip    []byte
	SubmissionID string
}

func downloadFromS3(key string, subID string, s3ChannelData chan S3ChannelData, process *sync.WaitGroup) {
	buff := &aws.WriteAtBuffer{}
	numBytes, err := downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(config.Bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		exitErrorf("problems downloading", err)
	}

	fmt.Println("Downloaded ", key, numBytes, "bytes")
	s3ChannelData <- S3ChannelData{ItemToZip: buff.Bytes(), SubmissionID: subID}
	process.Done()
}

func processMessages(w http.ResponseWriter, processComplete *sync.WaitGroup, s3ChannelData chan S3ChannelData) {

	// Loop over files, add them to the
	zipWriter := zip.NewWriter(w)
	// Register a custom Deflate compressor.
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.NoCompression)
	})

	processComplete.Add(1)
	fmt.Printf("Process has sugar\n")
	for s3Data := range s3ChannelData {
		fmt.Printf("\nFectech: %s\n", s3Data.SubmissionID)
		// We have to set a special flag so zip files recognize utf file names
		// See http://stackoverflow.com/questions/30026083/creating-a-zip-archive-with-unicode-filenames-using-gos-archive-zip
		h := &zip.FileHeader{
			Name:   s3Data.SubmissionID,
			Method: zip.Deflate,
			Flags:  0x800,
		}
		f, _ := zipWriter.CreateHeader(h)
		f.Write(s3Data.ItemToZip)

	}

	zipWriter.SetComment("--end--")
	zipWriter.Close()
	processComplete.Done()
	fmt.Printf("Proces dugger\n")
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get "ref" URL params
	campaignIds, ok := r.URL.Query()["campaignId"]
	if !ok || len(campaignIds) < 1 {
		http.Error(w, "S3 File Zipper. Pass ?ref= to use.", 500)
		return
	}
	campaignID := campaignIds[0]

	s3DataChannel := make(chan S3ChannelData, 5000)
	processComplete := sync.WaitGroup{}
	w.Header().Add("Content-Disposition", "attachment; filename=\"download.zip\"")
	w.Header().Add("Content-Type", "application/zip")
	go processMessages(w, &processComplete, s3DataChannel)

	s3Files := fetchS3Items(campaignID)
	processS3Items := sync.WaitGroup{}
	count := 0
	for _, file := range s3Files {
		if file.Path == "" {
			log.Printf("Missing path for file: %v", file)
			continue
		}
		processS3Items.Add(1)
		go downloadFromS3(file.Path, file.SubmissionID, s3DataChannel, &processS3Items)
		count = count + 1
		if count == 50 {
			processS3Items.Wait()
			count = 0
		}
	}

	processS3Items.Wait()
	log.Printf("\n\n%s\t%s\t DigDag to Download items from s3: %s", r.Method, r.RequestURI, time.Since(start))
	close(s3DataChannel)
	processComplete.Wait()
	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}

type s3Item struct {
	Path         string
	SubmissionID string
}

func awsDownloader() *s3manager.Downloader {
	ugcCert := util.UgcCert{}
	ugcHTTP := ugchttp.UgcHttp{Timeout: time.Duration(15 * time.Second),
		UgcCert: &ugcCert}
	ugcHTTP.InitClient(true)
	timeout := time.Duration(15 * time.Second)
	wh := wormhole.WormHole{UgcHttp: &ugcHTTP, Timeout: timeout}
	ugcAwsSession := ugcAws.AWS{WormHole: &wh, UgcCert: &ugcCert}
	sess := ugcAwsSession.AwsSession("")

	return s3manager.NewDownloader(sess)
}
func fetchS3Items(campaignID string) []s3Item {

	client := grab.NewClient()
	//client.HTTPClient.Transport.DisableCompression = true
	//u22417071
	req, _ := grab.NewRequest("/tmp", fmt.Sprintf("http://192.168.192.10:8080/export/campaign/%s/detailszip", campaignID))
	req.NoResume = true
	req.HTTPRequest.Header.Set("Accept", "application/zip")
	req.HTTPRequest.Header.Set("SSlClientCertSubject", "Email=kodjo.afriyie01@bbc.co.uk")

	resp := client.Do(req)

	t := time.NewTicker(time.Second)

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}
	t.Stop()

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)
	fmt.Printf("------------ FINISHED DOWNLOADING SUBMISSIONS ZIP FILE-----")

	zf, _ := zip.OpenReader(resp.Filename)
	defer zf.Close()
	defer os.Remove(resp.Filename)

	var s3Items []s3Item

	for _, file := range zf.File {
		fmt.Printf("=%s\n", file.Name)

		fc, err := file.Open()
		if err != nil {
			exitErrorf(fmt.Sprintf("Problems open file:%s", file.Name), err)
		}

		content, _ := ioutil.ReadAll(fc)
		fc.Close()

		fmt.Printf("%s\n\n", content) // file content
		cnt := string(content)
		for _, s := range strings.Split(strings.Replace(cnt, "\r\n", "\n", -1), "\n") {

			fileID := strings.TrimSpace(s)
			if len(fileID) > 0 {

				s3Item := s3Item{fmt.Sprintf("%s/%s/original", file.Name, fileID), file.Name}
				s3Items = append(s3Items, s3Item)
			}
		}
	}
	return s3Items

}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
