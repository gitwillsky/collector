package collector

import (
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector("http://www.baidu.com/index.html",
		1, 100,
		NewHttpDownloader(4*time.Second),
		NewHttpResolver())

	c.Start()
}
