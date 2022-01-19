package maingoroutine

import (
	"github.com/smallsung/gopkg/errors"
	"go.uber.org/zap"
)

type Service interface {
	Start() error
	Stop() error
}

func (g *Goroutine) RegisterService(service Service) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.state != initializingState {
		return errors.New("can't register service on running/stopped main goroutine ")
	}
	for _, srv := range g.services {
		if service == srv {
			return errors.Format("attempt to register service %g more than once", interfaceName(service))
		}
	}
	g.services = append(g.services, service)
	return nil
}

func (g *Goroutine) startServices(services []Service) ([]Service, error) {
	var started []Service
	for _, service := range services {
		name := interfaceName(service)
		fields := []zap.Field{zap.String("T", name)}
		g.logger.Info("starting service", fields...)
		if err := service.Start(); err != nil {
			g.logger.Warn("start service failure", fields...)
			return started, errors.Annotatef(err, "start service %s", name)
		}
		started = append(started, service)
		g.logger.Info("started service success", fields...)
	}
	return started, nil
}

func (g *Goroutine) stopServices(services []Service) (errs []error) {
	for i := len(services) - 1; i >= 0; i-- {
		service := services[i]
		name := interfaceName(service)
		fields := []zap.Field{zap.String("T", name)}
		g.logger.Info("stopping service", fields...)
		if err := service.Stop(); err != nil {
			errs = append(errs, errors.Annotatef(err, "stop service %s", name))
			g.logger.Warn("stop service failure", fields...)
			continue
		}
		g.logger.Info("stopped service success", fields...)
	}
	return errs
}
