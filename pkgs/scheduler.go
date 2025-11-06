package pkgs

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// AppScheduler 定时任务调度器
// 依赖注入：Logger、DB
// Start 方法启动定时任务

type Scheduler struct {
	Logger *zap.Logger
	DB     *sqlx.DB
}

func NewScheduler(logger *zap.Logger, db *sqlx.DB) *Scheduler {
	return &Scheduler{
		Logger: logger,
		DB:     db,
	}
}

func (s *Scheduler) Start() {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		s.Logger.Error("创建调度器失败", zap.Error(err))
		return
	}
	// 定时任务函数
	task := func() {
		if err := InitAdminRoot(s.DB, s.Logger); err != nil {
			s.Logger.Error("定时任务 InitAdminRoot 执行失败", zap.Error(err))
		} else {
			s.Logger.Info("定时任务 InitAdminRoot 执行成功")
		}
	}

	// 启动时立即执行一次
	task()

	_, jobErr := scheduler.NewJob(
		gocron.CronJob("*/5 * * * *", false),
		gocron.NewTask(task),
	)
	if jobErr != nil {
		s.Logger.Error("注册定时任务失败", zap.Error(jobErr))
		return
	}
	scheduler.Start()
	s.Logger.Info("定时任务 InitAdminRoot 已启动", zap.String("cron", "*/5 * * * *"))
}
