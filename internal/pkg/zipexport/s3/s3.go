package s3

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cavaliercoder/grab"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	ugcAws "github.com/NathanLewis/go-sqs-lambda-hello-world/internal/app/zipexport/aws"
	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/app/zipexport/wormhole"
	ugchttp "github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/http"
	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/util"
)

var config = util.Configuration{}

//S3Operations s3 operations
type Operations struct{}

//ChannelData data used by the channel
type ChannelData struct {
	ItemToZip    []byte
	SubmissionID string
}

var downloader = awsDownloader()

//DownloadFromS3 used to download items from s3
func (s3Operations *Operations) DownloadFromS3(key string, subID string, s3ChannelData chan ChannelData, process *sync.WaitGroup) {
	buff := &aws.WriteAtBuffer{}
	numBytes, err := downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(config.Bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		s3Operations.exitErrorf("problems downloading", err)
	}

	fmt.Println("Downloaded ", key, numBytes, "bytes")
	s3ChannelData <- ChannelData{ItemToZip: buff.Bytes(), SubmissionID: subID}
	process.Done()
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

//ProcessMessages used to process the data within the channel
func (s3Operations *Operations) ProcessMessages(w http.ResponseWriter, processComplete *sync.WaitGroup, s3ChannelData chan ChannelData) {

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

type s3Item struct {
	Path         string
	SubmissionID string
}

//FetchS3Items used to fetch items from s3
func (s3Operations *Operations) FetchS3Items(campaignID string) []s3Item {

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
			s3Operations.exitErrorf(fmt.Sprintf("Problems open file:%s", file.Name), err)
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

func (s3Operations *Operations) exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
