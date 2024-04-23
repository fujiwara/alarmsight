# alarmsight

alarmsight is a lambda function that queries the CloudWatch Insights API for logs and sends it to a Slack channel. It is intended to be used as a CloudWatch Alarm action.

## Usage

### Architecture

1. CloudWatch Alarm triggers the lambda function.
2. The lambda function queries logs with the CloudWatch Insights API.
   - The query is based on the alarm name.
3. The lambda function sends the query result to a Slack channel.

### Requirements

Environment variables:

- `SLACK_TOKEN`: Slack bot token
- `SLACK_CHANNEL`: Slack channel ID (not name)
- `QUERY_DURATION`: Duration of the query in seconds. (Default is 10 minutes)
   alarmsight queries logs that occurred within this duration.
- `QUERY_NAME_PREFIX`: Prefix of the query name. (Default is `alarmsight_`)
   alarmsight creates a query name with this prefix and the alarm name. For example, if the alarm name is `my-alarm`, the query name is `alarmsight_my-alarm`.

The environment variables can be loaded from the SSM Parameter Store.
You set the `SSM_PATH` environment variable to load the environment variables from the SSM Parameter Store.

For example, if you set `SSM_PATH` to `/alarmsight/`, alarmsight loads the following variables from SSM Parameter Store:
- `/alarmsight/SLACK_TOKEN`
- `/alarmsight/SLACK_CHANNEL`
- `/alarmsight/QUERY_DURATION`
- `/alarmsight/QUERY_NAME_PREFIX`

### Runtime

`provided.al2` or `provided.al2023` is supported.

alarmsight is written in Go. The release binary works as a lambda function's `bootstrap`.

You can rename the binary to `bootstrap` and create a zip file with the `bootstrap` file.

See also [examples](./examples) directory. The directory contains a sample `function.jsonnet` file for [lambroll](https://github.com/fujiwara/lambroll) to deploy a lambda function.

## LICENSE

MIT

## Author

Copyright (c) 2024 FUJIWARA Shunichiro
