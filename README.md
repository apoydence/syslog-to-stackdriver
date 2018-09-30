# syslog-to-stackdriver
Syslog Drain that stores logs to GCP's Stackdriver

### Log ID
Stackdriver has a concept called `LogID`. It is a way to logically organize logs. syslog-to-stackdriver will set the `LogID` of each message based on the last part of the URL path (e.g., `/some/path/MyLogID` results in `MyLogID`). If the path is empty, then it defaults to the configured value (which defaults to `syslog`).

### Deploying

syslog-to-stackdriver supports both Cloud Foundry and App Engine (standard environment) deploys. There is a script for both in the [scripts](https://github.com/apoydence/syslog-to-stackdriver/tree/master/scripts) directory. The App Engine method is preferred as it includes auto scaling and requires less configuration.
