package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/queue"
)

type Service struct {
	logger *slog.Logger
	queue  queue.Queue
	deploy deploy.Service
}

func New(logger *slog.Logger, q queue.Queue, d deploy.Service) *Service {
	return &Service{logger: logger, queue: q, deploy: d}
}

func (s *Service) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		job, err := s.queue.Dequeue(ctx, []queue.JobType{queue.JobTypeDeploy, queue.JobTypeCompile, queue.JobTypeReconcile}, 2*time.Second)
		if err != nil {
			if errors.Is(err, queue.ErrEmpty) {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			s.logger.Error("dequeue failed", "error", err)
			continue
		}
		s.handleJob(ctx, job)
	}
}

func (s *Service) handleJob(ctx context.Context, job queue.Job) {
	switch job.Type {
	case queue.JobTypeDeploy:
		runID, _ := job.Payload["run_id"].(string)
		if runID == "" {
			runID = job.ID
		}
		if _, err := s.deploy.ExecuteRun(ctx, runID); err != nil {
			s.logger.Error("execute deployment run", "run_id", runID, "error", err)
		}
	default:
		s.logger.Info("job type currently no-op in worker simulation", "type", job.Type, "job_id", job.ID)
	}
}
