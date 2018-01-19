package collector

import (
	"io/ioutil"
	"os"
)

type httpResolver struct {
}

func NewHttpResolver() *httpResolver {
	return &httpResolver{}
}

func (h *httpResolver) Resolve(data *Data) (*Address, error) {
	ioutil.WriteFile(data.Name, data.Content, os.ModePerm)
	return nil, nil
}
