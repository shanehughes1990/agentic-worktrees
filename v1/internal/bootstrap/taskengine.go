package bootstrap

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/infrastructure/observability"
	asynqengine "agentic-orchestrator/internal/infrastructure/queue/asynq"
	"fmt"
)

func bootstrapTaskEngine(config BaseConfig, observabilityPlatform *observability.Platform) (*taskengine.Scheduler, *asynqengine.Platform, error) {
	if config.TaskEngineBackend != "asynq" {
		return nil, nil, fmt.Errorf("unsupported task engine backend: %s", config.TaskEngineBackend)
	}

	entry := observabilityPlatform.ServiceEntry()
	if entry != nil {
		entry = entry.WithFields(map[string]any{"component": "taskengine", "backend": config.TaskEngineBackend})
	}

	platform := asynqengine.NewPlatform(asynqengine.Config{
		RedisAddress:  config.TaskEngineRedisAddress,
		RedisPassword: config.TaskEngineRedisPassword,
		RedisDatabase: config.TaskEngineRedisDatabase,
		Concurrency:   config.TaskEngineConcurrency,
	}, entry)

	policies := taskengine.DefaultPolicies()
	ingestionPolicy := policies[taskengine.JobKindIngestionAgent]
	ingestionPolicy.DefaultQueue = config.TaskEngineIngestionQueue
	policies[taskengine.JobKindIngestionAgent] = ingestionPolicy

	scheduler, err := taskengine.NewScheduler(platform, policies)
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap task engine scheduler: %w", err)
	}
	return scheduler, platform, nil
}
