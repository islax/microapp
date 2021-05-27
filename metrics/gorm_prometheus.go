package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/islax/microapp/config"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type Gormmetrics struct {
	*gorm.DB
	*DBStats
	refreshOnce, pushOnce sync.Once
	Labels                map[string]string
	collectors            []prometheus.Collector
	Config                *config.Config
}

func RegisterGormMetrics(db *gorm.DB, appConfig *config.Config) error {
	p := Gormmetrics{
		DB:     db,
		Config: appConfig,
	}
	p.DBStats = newStats(p.Labels)
	p.refreshOnce.Do(func() {
		go func() {
			for range time.Tick(time.Duration(p.Config.GetInt(config.EvSuffixForGormMetricsRefresh)) * time.Second) {
				p.refresh()
			}
		}()
	})

	return nil
}

func (p *Gormmetrics) refresh() {
	if db, err := p.DB.DB(); err == nil {
		p.DBStats.Set(db.Stats())
	} else {
		p.DB.Logger.Error(context.Background(), "gorm:prometheus failed to collect db status, got error: %v", err)
	}
}
