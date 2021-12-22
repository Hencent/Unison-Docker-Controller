package controller

import "testing"

func TestImageList(t *testing.T) {
	dc := getDockerControllerForTest(t)

	err := dc.ImageList()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
