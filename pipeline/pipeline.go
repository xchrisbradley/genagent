package pipeline

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"

	"encore.app/pipeline/activities"
	"encore.dev"
	activity "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	envName           = encore.Meta().Environment.Name
	pipelineTaskQueue = envName + "-pipeline"
)

var secrets struct {
	TemporalApiKey string
}

//encore:service
type Service struct {
	client client.Client
	worker worker.Worker
}

// initService is automatically called by Encore when the service starts up.
func initService() (*Service, error) {
	opts := client.Options{
		HostPort:  cfg.TemporalServer,
		Namespace: cfg.TemporalNamespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: func() *tls.Config {
				if encore.Meta().Environment.Cloud == "local" {
					return nil // Disable TLS for local development
				}
				return &tls.Config{}
			}(),
			DialOptions: []grpc.DialOption{
				grpc.WithUnaryInterceptor(
					func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						return invoker(
							metadata.AppendToOutgoingContext(ctx, "temporal-namespace", cfg.TemporalNamespace),
							method,
							req,
							reply,
							cc,
							opts...,
						)
					},
				),
			},
		},
		Credentials: func() client.Credentials {
			if encore.Meta().Environment.Cloud == "local" {
				return nil // No auth needed for local development
			}
			return client.NewAPIKeyStaticCredentials(secrets.TemporalApiKey)
		}(),
	}
	c, err := client.Dial(opts)
	if err != nil {
		return nil, fmt.Errorf("create temporal client: %v", err)
	}

	w := worker.New(c, pipelineTaskQueue, worker.Options{})

	// Register workflow executor
	w.RegisterWorkflowWithOptions(
		ExecutePipeline,
		workflow.RegisterOptions{
			Name: "ExecutePipeline",
		},
	)

	// Register system activities with options
	httpActivity := activities.NewHTTPActivity()
	w.RegisterActivityWithOptions(
		httpActivity.Execute,
		activity.RegisterOptions{
			Name: "HTTPActivity",
		},
	)

	err = w.Start()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker: %v", err)
	}
	return &Service{client: c, worker: w}, nil
}

func (s *Service) Close() {
	s.worker.Stop()
	s.client.Close()
}
