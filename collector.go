package collector

import (
	"time"
	"sync"
	"log"
	"sync/atomic"
)

type Downloader interface {
	Download(address *Address) (*Data, error)
}

type Resolver interface {
	Resolve(data *Data) (*Address, error)
}

type collector struct {
	startPoint  string
	addressChan chan *Address
	dataChan    chan *Data
	deadLine    time.Duration
	wg          sync.WaitGroup
	downloader  Downloader
	resolver    Resolver
	downloaded  uint32
	limit       uint32
	maxDepth    uint32
}

type Data struct {
	DataType string
	Name     string
	Content  []byte
	Depth    uint32
}

type Address struct {
	URL   string
	Depth uint32
}

func NewCollector(startPoint string, maxDepth, maxCount uint32, downloader Downloader, resolver Resolver) *collector {
	c := &collector{
		startPoint:  startPoint,
		addressChan: make(chan *Address, 10),
		dataChan:    make(chan *Data, 10),
		deadLine:    10 * time.Second,
		downloader:  downloader,
		resolver:    resolver,
		limit:       maxCount,
		maxDepth:    maxDepth,
	}

	c.addressChan <- &Address{URL: startPoint, Depth: 0}
	return c
}

func (c *collector) Start() {
	c.wg.Add(2)
	go c.startDownload()
	go c.startParse()
	c.wg.Wait()
}

func (c *collector) startDownload() {
	defer c.wg.Done()
	for {
		time.Sleep(time.Second)
		select {
		case newAddress := <-c.addressChan:
			// 达到最大深度
			if (newAddress.Depth >= c.maxDepth) {
				break
			}
			// 达到最大数量
			if (atomic.LoadUint32(&c.downloaded) == c.limit) {
				break
			}

			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				data, err := c.downloader.Download(newAddress)
				if err != nil {
					log.Print(err.Error())
					return
				}

				atomic.AddUint32(&c.downloaded, 1)
				c.dataChan <- data
			}()
		case <-time.After(c.deadLine):
			log.Println("downloader timeout")
			return
		}
	}
}

func (c *collector) startParse() {
	defer c.wg.Done()
	for {
		select {
		case data := <-c.dataChan:
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()

				newAddress, err := c.resolver.Resolve(data)
				if err != nil {
					log.Println(err.Error())
					return
				}
				if (newAddress != nil) {
					c.addressChan <- newAddress
				}
			}()
		case <-time.After(c.deadLine):
			log.Println("resolver timeout")
			return
		}
	}
}
