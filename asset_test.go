package main

import (
	"testing"
)

func Test_ReadFromString(t *testing.T) {
	var asset Asset
	const message = `
	<simpleAsset activityId="fd105e8a-b86c-4509-9cff-528c7e2160eb">
		<uri>s3://source_bucket/key.mp4</uri>
	</simpleAsset>`
	err := asset.readFromString(message)
	if err != nil {
		t.Errorf("failed to parse xml: %v", message)
	}
	const expectedUri = "s3://source_bucket/key.mp4"
	if expectedUri != asset.Uri {
		t.Errorf("Expected %v, got %v", expectedUri, asset.Uri)
	}
}

func Test_AssetToMap(t *testing.T) {
	var asset Asset

	const uri = "s3://source_bucket/key.mp4"
	asset.Uri = uri
	asset.ActivityId = "fd105e8a-b86c-4509-9cff-528c7e2160eb"

	actualMap := asset.toMap()
	actualUri := actualMap["uri"]
	if uri != actualUri {
		t.Errorf("Expected %v, got %v", uri, actualUri)
	}
}