package youtube

import (
	"fmt"
	"os"
	"testing"
)

func TestService_SearchAndDownload(t *testing.T) {
	svc := NewService(480)
	media, err := svc.SearchAndDownload("rickroll")
	if err != nil {
		t.Error("expected no error, got", err.Error())
	}
	fmt.Println(media.URL)
	fmt.Println(media.Title)
	fmt.Println(media.Thumbnail)
	fmt.Println(media.FilePath)
	_ = os.Remove(media.FilePath)
}
