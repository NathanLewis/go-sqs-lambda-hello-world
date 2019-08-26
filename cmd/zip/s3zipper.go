package main

import (
	"archive/zip"
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
	"time"

	"github.com/cavaliercoder/grab"
	"golang.org/x/net/http2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"go-sqs-lambda-hello-world/internal/pkg/zipexport/util"
	ugcAws "go-sqs-lambda-hello-world/internal/app/zipexport/aws"
	"go-sqs-lambda-hello-world/internal/app/zipexport/wormhole"
	ugchttp "go-sqs-lambda-hello-world/internal/pkg/zipexport/http"
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

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}

func initAwsBucket() {

	downloader = awsDownloader()
	downloader.Concurrency = 1

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

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	health, ok := r.URL.Query()["health"]
	if len(health) > 0 {
		fmt.Fprintf(w, "OK")
		return
	}

	// Get "ref" URL params
	campaignIds, ok := r.URL.Query()["campaignId"]
	if !ok || len(campaignIds) < 1 {
		http.Error(w, "S3 File Zipper. Pass ?ref= to use.", 500)
		return
	}
	campaignID := campaignIds[0]

	// Start processing the response
	w.Header().Add("Content-Disposition", "attachment; filename=\"download.zip\"")
	w.Header().Add("Content-Type", "application/zip")

	// Loop over files, add them to the
	zipWriter := zip.NewWriter(w)

	s3Files := fetchS3Items(campaignID)

	for _, file := range s3Files {

		if file.Path == "" {
			log.Printf("Missing path for file: %v", file)
			continue
		}

		pr, pw := io.Pipe()
		numBytes, err := downloader.Download(FakeWriterAt{pw},
			&s3.GetObjectInput{
				Bucket: aws.String(config.Bucket),
				Key:    aws.String(file.Path),
			})

		pw.Close()
		if err != nil {
			exitErrorf("problems downloading", err)
		}

		fmt.Println("Downloaded", file.Path, numBytes, "bytes")

		// We have to set a special flag so zip files recognize utf file names
		// See http://stackoverflow.com/questions/30026083/creating-a-zip-archive-with-unicode-filenames-using-gos-archive-zip
		h := &zip.FileHeader{
			Name:   file.SubmissionID,
			Method: zip.Deflate,
			Flags:  0x800,
		}

		f, _ := zipWriter.CreateHeader(h)
		io.Copy(f, pr)
		pr.Close()

	}

	zipWriter.Close()
	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}

type s3Item struct {
	Path         string
	SubmissionID string
}

func awsDowloader()  *s3manager.Downloader {
	ugcCert := util.UgcCert{}
	ugcHTTP := ugchttp.UgcHttp{Timeout: time.Duration(15 * time.Second),
		UgcCert: &ugcCert}
	ugcHttp.InitClient(true)
	timeout := time.Duration(15 * time.Second)
	wh := wormhole.WormHole{UgcCert: &ugcCert, UgcHttp: &ugcHTTP, Timeout: timeout}
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
	defer t.Stop()
	fmt.Sprintf("------------ FINISHED DOWNLOADING SUBMISSIONS ZIP FILE-----")

	zf, _ := zip.OpenReader(resp.Filename)
	defer zf.Close()

	var s3Items []s3Item

	for _, file := range zf.File {
		fmt.Printf("=%s\n", file.Name)
		content := readAll(file)
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

// readAll is a wrapper function for ioutil.ReadAll. It accepts a zip.File as
// its parameter, opens it, reads its content and returns it as a byte slice.
func readAll(file *zip.File) []byte {
	fc, _ := file.Open()
	defer fc.Close()
	content, _ := ioutil.ReadAll(fc)
	return content
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
