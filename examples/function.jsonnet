{
  FunctionName: 'alarmsight',
  Handler: 'index.handler',
  MemorySize: 128,
  Role: 'arn:aws:iam::314472643515:role/alarmsight-lambda',
  Runtime: 'provided.al2023',
  Timeout: 30,
  Environment: {
    Variables: {
      SLACK_CHANNEL: 'CDUMMY',
      SLACK_TOKEN: 'xoxb-dummy',
      SKIP_POST: 'true', // for debugging
      LOG_JSON: 'true',
    }
  }
}
