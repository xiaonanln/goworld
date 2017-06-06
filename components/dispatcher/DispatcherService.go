package main

import "github.com/xiaonanln/vacuum/config"

type DispatcherService struct {
}

func newDispatcherService(cfg *config.DispatcherConfig) *DispatcherService {
	return &DispatcherService{}
}
