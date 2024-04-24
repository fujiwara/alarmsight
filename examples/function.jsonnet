{
  FunctionName: 'alarmsight',
  Handler: 'bootstrap',
  MemorySize: 128,
  Role: 'arn:aws:iam::{{ env `AWS_ACCOUNT_ID` }}:role/alarmsight-lambda',
  Runtime: 'provided.al2023',
  Timeout: 30,
  LoggingConfig: {
    LogGroup: '/aws/lambda/alarmsight',
    ApplicationLogLevel: 'INFO',
    LogFormat: 'JSON',
    SystemLogLevel: 'INFO',
  },
  Environment: {
    Variables: {
      SLACK_CHANNEL: 'CDUMMY',
      SLACK_TOKEN: 'xoxb-dummy',
      SKIP_POST: 'true', // for debugging
      LOG_JSON: 'true',
    }
  }
}
