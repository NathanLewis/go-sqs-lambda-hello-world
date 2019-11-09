package zipexport

import (
	"log"
	"net/http"
	"sync"

	ugcS3 "github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/s3"
)

//ZipProcessor operations
type ZipProcessor struct{}

//Process used to process items to be zipped
func (zipProcessor ZipProcessor) Process(w http.ResponseWriter, campaignID string) {

	s3Operations := ugcS3.Operations{}
	s3DataChannel := make(chan ugcS3.ChannelData, 5000)
	processComplete := sync.WaitGroup{}
	go s3Operations.ProcessMessages(w, &processComplete, s3DataChannel)

	s3Files := s3Operations.FetchS3Items(campaignID)
	processS3Items := sync.WaitGroup{}
	count := 0
	for _, file := range s3Files {
		if file.Path == "" {
			log.Printf("Missing path for file: %v", file)
			continue
		}
		processS3Items.Add(1)
		go s3Operations.DownloadFromS3(file.Path, file.SubmissionID, s3DataChannel, &processS3Items)
		count = count + 1
		if count == 50 {
			processS3Items.Wait()
			count = 0
		}
	}

	processS3Items.Wait()
	close(s3DataChannel)
	processComplete.Wait()
}
