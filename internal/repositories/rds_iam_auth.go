package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	rdsauth "github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/jackc/pgx/v5"
)

const rdsIAMAuthTokenTTL = 15 * time.Minute

type rdsIAMTokenProvider struct {
	endpoint    string
	region      string
	username    string
	credentials aws.CredentialsProvider
	refreshLead time.Duration

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func newRDSIAMTokenProvider(ctx context.Context, connConfig *pgx.ConnConfig, awsRegion string, refreshLead time.Duration) (*rdsIAMTokenProvider, error) {
	if connConfig == nil {
		return nil, fmt.Errorf("postgres conn config is required")
	}

	if awsRegion == "" {
		return nil, fmt.Errorf("AWS region is required for RDS IAM auth")
	}

	if connConfig.Host == "" || connConfig.Port == 0 {
		return nil, fmt.Errorf("postgres host and port are required for RDS IAM auth")
	}

	if connConfig.User == "" {
		return nil, fmt.Errorf("postgres user is required for RDS IAM auth")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("load AWS config for RDS IAM auth: %w", err)
	}

	return &rdsIAMTokenProvider{
		endpoint:    fmt.Sprintf("%s:%d", connConfig.Host, connConfig.Port),
		region:      awsRegion,
		username:    connConfig.User,
		credentials: awsCfg.Credentials,
		refreshLead: refreshLead,
	}, nil
}

func (p *rdsIAMTokenProvider) ApplyToken(ctx context.Context, connConfig *pgx.ConnConfig) error {
	if connConfig == nil {
		return fmt.Errorf("postgres conn config is required")
	}

	token, err := p.getToken(ctx)
	if err != nil {
		return err
	}

	connConfig.Password = token
	return nil
}

func (p *rdsIAMTokenProvider) getToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token != "" && time.Until(p.expiresAt) > p.refreshLead {
		return p.token, nil
	}

	token, err := rdsauth.BuildAuthToken(ctx, p.endpoint, p.region, p.username, p.credentials)
	if err != nil {
		return "", fmt.Errorf("build RDS IAM auth token: %w", err)
	}

	p.token = token
	p.expiresAt = time.Now().Add(rdsIAMAuthTokenTTL)

	return p.token, nil
}
