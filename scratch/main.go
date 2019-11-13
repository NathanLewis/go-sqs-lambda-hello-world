package main

import (
    "fmt"
    "encoding/xml"
)

func main() {
    data:=`
    <?xml version="1.0" encoding="UTF-8"?>
    <simpleAsset activityId="fd105e8a-b86c-4509-9cff-528c7e2160eb">
        <uri>s3://source_bucket/key.mp4</uri>
    </simpleAsset>
    `

    type Uri struct {
        XMLName xml.Name `xml:"uri"`
        Path string
    }

    type SimpleAsset struct {
        XMLName xml.Name `xml:"simpleAsset"`
        ActivityId string  `xml:"activityId,attr"`  // notice the capitalized field name here and the `xml:"app_name,attr"`
        Uri Uri
    }

    var asset SimpleAsset
    err := xml.Unmarshal([]byte(data), &asset)
    if err != nil {
        fmt.Printf("error: %v", err)
        return
    }
    fmt.Printf("asset ID:: %q\n", asset.ActivityId)       // the corresponding changes here for App
    fmt.Printf("asset location:: %q\n", asset.Uri.Path)   // the corresponding changes here for App
}
