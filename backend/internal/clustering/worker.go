package clustering

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/client"
	"prod-pobeda-2026/internal/usecase"
)

const defaultMinClusterSize = 3

type Worker struct {
	brandUseCase   *usecase.BrandUseCase
	mlClient       client.SentimentMLClient
	clusteringRepo usecase.ClusteringRepository
	interval       time.Duration
	running        atomic.Bool
}

func NewWorker(
	brandUseCase *usecase.BrandUseCase,
	mlClient client.SentimentMLClient,
	clusteringRepo usecase.ClusteringRepository,
	interval time.Duration,
) *Worker {
	return &Worker{
		brandUseCase:   brandUseCase,
		mlClient:       mlClient,
		clusteringRepo: clusteringRepo,
		interval:       interval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tryProcess(ctx)
		}
	}
}

func (w *Worker) tryProcess(ctx context.Context) {
	if !w.running.CompareAndSwap(false, true) {
		logrus.Info("clustering worker already running, skipping cycle")
		return
	}
	defer w.running.Store(false)

	logrus.Info("clustering worker: starting cycle")
	if err := w.process(ctx); err != nil {
		logrus.WithError(err).Error("clustering worker: cycle failed")
	}
}

func (w *Worker) process(ctx context.Context) error {
	brands, _, err := w.brandUseCase.List(ctx, 100, 0)
	if err != nil {
		return err
	}

	logrus.Infof("clustering worker: processing %d brands", len(brands))

	for _, brand := range brands {
		if err := w.processBrand(ctx, brand.ID, brand.Name); err != nil {
			logrus.WithError(err).Warnf("clustering worker: brand %s failed, skipping", brand.Name)
		}
	}

	return nil
}

func (w *Worker) processBrand(ctx context.Context, brandID uuid.UUID, brandName string) error {
	items, err := w.clusteringRepo.GetItemsByBrandID(ctx, brandID)
	if err != nil {
		return err
	}

	if len(items) < defaultMinClusterSize {
		logrus.Infof("clustering worker: brand %s has %d items, skipping (min %d)", brandName, len(items), defaultMinClusterSize)
		return nil
	}

	texts := make([]string, len(items))
	textToID := make(map[string]uuid.UUID, len(items))
	for i, item := range items {
		texts[i] = item.Text
		textToID[item.Text] = item.ID
	}

	result, err := w.mlClient.Cluster(texts, defaultMinClusterSize)
	if err != nil {
		return err
	}

	assignments := make(map[uuid.UUID]*int, len(items))

	for i := range result.Clusters {
		cluster := &result.Clusters[i]
		label := cluster.ClusterID
		for _, msg := range cluster.Messages {
			if itemID, ok := textToID[msg]; ok {
				labelCopy := label
				assignments[itemID] = &labelCopy
			}
		}
	}

	for _, msg := range result.Noise {
		if itemID, ok := textToID[msg]; ok {
			assignments[itemID] = nil
		}
	}

	if err := w.clusteringRepo.UpdateClusterLabels(ctx, brandID, assignments); err != nil {
		return err
	}

	logrus.Infof("clustering worker: brand %s — %d clusters, %d noise, %d items updated",
		brandName, result.NumClusters, result.NumNoise, len(assignments))

	return nil
}
