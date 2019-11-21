package main

import (
	"encoding/xml"
	"fmt"
)

type Asset struct {
	XMLName    xml.Name `xml:"simpleAsset"`
	ActivityId string   `xml:"activityId,attr"`
	// notice the capitalized field ActivityId here and the `xml:"app_name,attr"`
	Uri        string `xml:"uri"`
}

func (asset *Asset) readFromString(message string) error {
	err := xml.Unmarshal([]byte(message), asset)
	return err
}

func (asset Asset) printFields() {
	fmt.Printf("asset ID:: %q\n", asset.ActivityId)
	fmt.Printf("asset location:: %q\n", asset.Uri)
}

func (asset Asset) toMap() (map[string]interface{}) {
	result := make(map[string]interface{})
	result["uri"] = asset.Uri
	result["activityId"] = asset.ActivityId

	return result
}
