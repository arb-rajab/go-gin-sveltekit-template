# Optional module: serverless example (AWS Lambda + SAM)

A standalone Lambda function behind an API Gateway HTTP API, built with the AWS
SAM CLI. It is a separate Go module (its own `go.mod`) and a separate
deployment target: it does not run inside the backend. Replace `handler.go`
with your own function, or delete this directory.

## Build, test, deploy

```
go vet ./... && go test ./... -race
sam build
sam validate --lint
sam deploy --guided
```

`sam local invoke` needs a container runtime and pulls an image; it has not been
exercised here. `events/` holds sample payloads for it:

```
sam local invoke GreetFunction --event events/valid.json
```

CI builds the function, runs its tests, and runs `sam validate --lint`; it never
deploys and needs no AWS credentials.
