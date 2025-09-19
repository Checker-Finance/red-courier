package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"red-courier/internal/config"
	"red-courier/internal/db"
	"red-courier/internal/redis"
	"red-courier/internal/task"
)

type Scheduler struct {
	cron    *cron.Cron
	tasks   []*task.Task
	context context.Context
}

func NewScheduler(ctx context.Context, cfg *config.Config, db *db.Database, redis *redis.RedisClient) (*Scheduler, error) {
	s := &Scheduler{
		cron:    cron.New(),
		tasks:   []*task.Task{},
		context: ctx,
	}

	for _, tcfg := range cfg.Tasks {
		t, err := task.NewTask(tcfg, db, redis)
		if err != nil {
			return nil, err
		}
		s.tasks = append(s.tasks, t)

		schedule := tcfg.Schedule
		if schedule == "" {
			schedule = "@every 5m"
		}

		log.Printf("Scheduling task %s to run %s", tcfg.Name, schedule)
		_, err = s.cron.AddFunc(schedule, func(taskToRun *task.Task) func() {
			return func() {
				ctx, cancel := context.WithTimeout(s.context, 1*time.Minute)
				defer cancel()

				log.Printf("Running scheduled task for: %s (schedule: %s)", taskToRun.Config.Table, schedule)
				if err := taskToRun.Run(ctx); err != nil {
					log.Printf("Error in task %s: %v", taskToRun.Config.Table, err)
				}
			}
		}(t))

		if err != nil {
			return nil, fmt.Errorf("failed to schedule task %s: %w", tcfg.Table, err)
		}
	}

	return s, nil
}

func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")
	s.cron.Start()
	select {}
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.cron.Stop()
}
