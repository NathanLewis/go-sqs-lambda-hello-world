package wormhole

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	ugchttp "github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/http"
	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/util"
)

//WormHole used ot connect to wormhole
type WormHole struct {
	Timeout time.Duration
	UgcHttp *ugchttp.UgcHttp
}

func (wormhole *WormHole) dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, wormhole.Timeout)
}

//WormholeResponse response from wormhole
type WormholeResponse struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      string `json:"expiration"`
	AssumedRole     string `json:"assumedRole"`
}

//SessionInfo used to connect to aws
func (wormHole *WormHole) SessionInfo() (wormHoleResponse WormholeResponse) {
	config := util.Configuration{}
	config = config.Config()
	wh := fmt.Sprintf("https://wormhole.api.bbci.co.uk/account/%s/credentials", config.AwsAccountID)
	body, err := wormHole.UgcHttp.FetchItemWormHole(wh, make(map[string]string))

	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &wormHoleResponse)
	return

}
