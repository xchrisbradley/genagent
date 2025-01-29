package agent

import (
	"context"
	"crypto/tls"
	"fmt"

	"encore.app/agent/activities/email/consent"
	"encore.app/agent/activities/email/reminder"
	"encore.app/agent/activities/email/research"
	"encore.app/agent/workflows"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"encore.dev"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Use an environment-specific task queue so we can use the same
// Temporal Cluster for all cloud environments.
var (
	envName        = encore.Meta().Environment.Name
	agentTaskQueue = envName + "-agent"
)

// GetTaskQueue returns the task queue name
func GetTaskQueue() string {
	return agentTaskQueue
}

// Client returns the temporal client
func (s *Service) Client() client.Client {
	return s.client
}

var secrets struct {
	TemporalApiKey string
}

//encore:service
type Service struct {
	client client.Client
	worker worker.Worker
}

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

	// Initialize activities
	researchActivity := research.NewActivity(nil)
	consentActivity := consent.NewActivity(nil)
	reminderActivity := reminder.NewActivity(nil)

	w := worker.New(c, agentTaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.Agent)
	w.RegisterActivity(researchActivity.SendResearchRequest)
	w.RegisterActivity(consentActivity.SendConsentEmail)
	w.RegisterActivity(reminderActivity.SendReminder)

	err = w.Start()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker: %v", err)
	}
	return &Service{client: c, worker: w}, nil
}

func (s *Service) Shutdown(force context.Context) {
	s.client.Close()
	s.worker.Stop()
}
