package alarmsight

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/slack-go/slack"
)

type CLI struct {
	SlackToken      string        `env:"SLACK_TOKEN" help:"Slack token" required:"true"`
	SlackChannel    string        `env:"SLACK_CHANNEL" help:"Slack Channel ID" required:"true"`
	QueryDuration   time.Duration `env:"QUERY_DURATION" help:"Duration of query"  default:"10m"`
	QueryNamePrefix string        `env:"QUERY_NAME_PREFIX" default:"alarmsight_"`
}

func NewCLI() *CLI {
	app := &CLI{}
	kong.Parse(app)
	return app
}

func (c *CLI) Handler(ctx context.Context, payload Payload) (struct{}, error) {
	err := c.process(ctx, payload)
	if err != nil {
		slog.Error("failed to process", "error", err)
	}
	return struct{}{}, err
}

// handler is the entry point for the Lambda function.
func (c *CLI) process(ctx context.Context, payload Payload) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return fmt.Errorf("failed to load aws credentials %w", err)
	}
	alarmName, state, err := ParsePayload(payload)
	if err != nil {
		return fmt.Errorf("failed to parse payload %w", err)
	}
	slog.Info("starting process", "alarm_name", alarmName, "alarm_state", state)
	if state != "ALARM" {
		slog.Info("state is not ALARM, skip", "alarm_name", alarmName, "alarm_state", state)
		return nil
	}

	svc := cloudwatchlogs.NewFromConfig(cfg)
	queryDef, err := c.findQueryByAlarmName(ctx, svc, alarmName)
	if err != nil {
		slog.Error("failed to find query definition", "error", err, "alarm_name", alarmName)
		return fmt.Errorf("failed to find query definition %w", err)
	}
	if queryDef == nil {
		slog.Info("query definition not found, skip", "alarm_name", alarmName)
		return nil
	}

	slog.Info("found queryDefinition",
		"query_definition_id", *queryDef.QueryDefinitionId,
		"query_name", *queryDef.Name,
	)
	results, err := c.doQueryResults(ctx, svc, queryDef)
	if err != nil {
		return fmt.Errorf("failed to do query results %w", err)
	}

	if len(results) == 0 {
		slog.Warn("no results found", "alarm_name", alarmName)
		return fmt.Errorf("no results found. retrying")
	}

	for _, line := range results {
		slog.Info("result", "record", line)
	}

	return c.postToSlack(ctx, *queryDef.Name, results)
}

// parsePayload parses the payload of a CloudWatch alarm and returns the name and state.
func ParsePayload(p Payload) (string, string, error) {
	r, err := arn.Parse(p.AlarmArn)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse resource ARN %s %w", p.AlarmArn, err)
	}
	if r.Service != "cloudwatch" || !strings.HasPrefix(r.Resource, "alarm:") {
		return "", "", fmt.Errorf("unexpected resource ARN %s", p.AlarmArn)
	}
	if p.AlarmData.AlarmName == "" {
		return "", "", fmt.Errorf("alarm name is empty")
	}
	if p.AlarmData.State.Value == "" {
		return "", "", fmt.Errorf("alarm state is empty")
	}
	return p.AlarmData.AlarmName, p.AlarmData.State.Value, nil
}

// findQueryByAlarmName finds a query definition by alarm name.
func (c *CLI) findQueryByAlarmName(ctx context.Context, svc *cloudwatchlogs.Client, alarmName string) (*types.QueryDefinition, error) {
	out, err := svc.DescribeQueryDefinitions(ctx, &cloudwatchlogs.DescribeQueryDefinitionsInput{
		QueryDefinitionNamePrefix: aws.String(c.QueryNamePrefix + alarmName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe query definitions %w", err)
	}
	if len(out.QueryDefinitions) == 0 {
		slog.Info("no query definitions found for alarm", "alarm_name", alarmName)
		return nil, nil
	}
	return &out.QueryDefinitions[0], nil
}

// doQueryResults executes a query and returns the results.
func (c *CLI) doQueryResults(ctx context.Context, svc *cloudwatchlogs.Client, queryDef *types.QueryDefinition) ([]string, error) {
	now := time.Now().Truncate(time.Second)
	start := now.Add(-c.QueryDuration).Truncate(time.Second)
	slog.Info("executing query", "query_string", *queryDef.QueryString, "start", start, "end", now)
	out, err := svc.StartQuery(ctx, &cloudwatchlogs.StartQueryInput{
		QueryString:   queryDef.QueryString,
		StartTime:     aws.Int64(start.Unix()),
		EndTime:       aws.Int64(now.Unix()),
		LogGroupNames: queryDef.LogGroupNames,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start query %w", err)
	}
	queryId := *out.QueryId
	slog.Info("query started", "query_id", queryId)
	ticker := time.NewTicker(1 * time.Second)
	var results []string
WAITER:
	for range ticker.C {
		out, err := svc.GetQueryResults(ctx, &cloudwatchlogs.GetQueryResultsInput{
			QueryId: &queryId,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get query results %w", err)
		}
		switch out.Status {
		case types.QueryStatusComplete:
			slog.Info("query completed", "query_id", queryId, "status", out.Status, "statistics", out.Statistics)
			for _, result := range out.Results {
				for _, r := range result {
					if aws.ToString(r.Field) == "@message" {
						results = append(results, aws.ToString(r.Value))
					}
				}
			}
			break WAITER
		case types.QueryStatusFailed, types.QueryStatusCancelled, types.QueryStatusTimeout:
			return nil, fmt.Errorf("failed to query: %s", out.Status)
		case types.QueryStatusRunning:
			slog.Info("query running...", "query_id", queryId, "status", out.Status)
			continue
		default:
			return nil, fmt.Errorf("unexpected query status: %s", out.Status)
		}
	}
	return results, nil
}

func (c *CLI) postToSlack(ctx context.Context, queryName string, results []string) error {
	client := slack.New(c.SlackToken)
	body := strings.Join(results, "\n")
	now := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.txt", queryName, now)
	slog.Info("posting to slack", "channel", c.SlackChannel, "filename", filename, "size", len(body))
	_, err := client.UploadFileV2Context(ctx, slack.UploadFileV2Parameters{
		Channel:  c.SlackChannel,
		Filename: filename,
		Content:  body,
		FileSize: len(body),
	})
	return fmt.Errorf("failed to upload file to slack %w", err)
}

/*
Payload represents the payload of a CloudWatch Alarm.
https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/AlarmThatSendsEmail.html#Lambda-action-payload
*/
type Payload struct {
	AccountID string `json:"accountId"`
	AlarmArn  string `json:"alarmArn"`
	AlarmData struct {
		AlarmName     string `json:"alarmName"`
		Configuration struct {
			Description string `json:"description"`
			Metrics     []any  `json:"metrics"`
		} `json:"configuration"`
		State struct {
			Reason    string `json:"reason"`
			Timestamp string `json:"timestamp"`
			Value     string `json:"value"`
		} `json:"state"`
	} `json:"alarmData"`
	Region string `json:"region"`
	Source string `json:"source"`
	Time   string `json:"time"`
}
