package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	TaskSendVerifyEmail = "task:send_verify_email"
)

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

// DistributeTaskSendVerifyEmail create a new task and send the task to Redis queue
func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	//convert payload to json
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload : %w", err)
	}
	//create a new task with TypeName which is asynq to recognize what kind of task that it is distributing or processing
	//with payload
	//with options which allow us to control how the task distributed,ran,retried.
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	//send the task to redis queue
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task to redis : %w", err)
	}
	log.Info().Str("task type name", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

// ProcessTaskSendVerifyEmail get the task and process it
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	//convert json payload to byte payload
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		//use asynq.SkipRetry to tell redis ???
		return fmt.Errorf("failed to unmarshal payload : %s", asynq.SkipRetry)
	}
	//get the user in store
	user, err := processor.Store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist in db : %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user :%w", asynq.SkipRetry)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
