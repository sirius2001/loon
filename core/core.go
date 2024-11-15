package core

import (
	"github.com/sirius2001/loon/config"
	"github.com/sirius2001/loon/pkg/db"
	"github.com/sirius2001/loon/pkg/grpc"
	"github.com/sirius2001/loon/pkg/kaf"
	"github.com/sirius2001/loon/pkg/log"
	"github.com/sirius2001/loon/pkg/web"
)

type Core struct {
	services []ServiceInner
}

func NewCore(confPath string) (*Core, error) {
	if err := config.LoadConfig(confPath); err != nil {
		return nil, err
	}

	log.SetupLogger(log.Config{
		Dir:      config.Conf().Dir,
		Level:    config.Conf().Level,
		Duration: config.Conf().Duration,
		MaxAge:   config.Conf().MaxAge,
		MaxSize:  config.Conf().MaxSize,
	})

	if config.Conf().DB.Enable {
		if err := db.NewDB(config.Conf().Merge, config.Conf().DSN); err != nil {
			log.Error("NewCore", "err", err)
		}
	}

	var services []ServiceInner
	if config.Conf().Web.Enable {
		webService, err := web.NewWebService(config.Conf().Web.Addr)
		if err != nil {
			log.Error("web service start failed", "err", err)
		} else {
			services = append(services, webService)
		}
	}

	if config.Conf().GRPC.Enable {
		grpcService, err := grpc.NewRPCServer(config.Conf().GRPC.Addr)
		if err != nil {
			log.Error("grpc service start failed", "err", err)
		} else {
			services = append(services, grpcService)
		}
	}

	if err := kaf.NewProducer(); err != nil {
		log.Error("connect to kafka error", "err", err)
		return nil, err
	}

	return &Core{
		services: services,
	}, nil
}

func (c *Core) Run() {
	for _, service := range c.services {
		go func() {
			if err := service.StartService(); err != nil {
				panic(err)
			}
		}()
		log.Info("service start suceessfully", "service", service.ServiceName(), "addr", service.ServiceAddr())
	}
}

func (c *Core) Stop() {
	for _, service := range c.services {
		if err := service.StopService(); err != nil {
			log.Error("stop service with err", "err", err)
		}
	}
}
