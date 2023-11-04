package worker

import (
	"context"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

// TaskProcessor define all method to process task
type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor provides all method to process task,it is real implement of TaskProcessor
type RedisTaskProcessor struct {
	Server *asynq.Server
	Store  db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Err(err).Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("process failed ")
			}),
			Logger: NewLogger(),
		},
	)
	return &RedisTaskProcessor{
		Server: server,
		Store:  store,
	}
}
func (processor *RedisTaskProcessor) Start() error {
	// create a new mux , likes http server, can use this mux to register every task and it's handler
	mux := asynq.NewServeMux()
	// register the task type name and it's handler func  to mux
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	// start the server
	return processor.Server.Start(mux)
}
