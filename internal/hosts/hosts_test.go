package hosts

import (
	"fmt"
	"testing"
)

func TestGetHostInfo(t *testing.T) {
	info, err := GetHostInfo()
	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Println(info)
}
