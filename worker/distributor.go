package worker

import (
	"context"
	"github.com/hibiken/asynq"
)

// TaskDistributor defines all method to distribute the task
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

// RedisTaskDistributor provides all function to distribute the task
// RedisTaskDistributor is the real implement of TaskDistributor interface
type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewRedisTaskDistributor create a new RedisTaskDistributor
// include all function of TaskDistributor interface
func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
