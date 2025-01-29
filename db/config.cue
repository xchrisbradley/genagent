package agent

TemporalServer: string | *"localhost:7233"
TemporalNamespace: string | *"default"

// Use cloud settings for any non-local environment
if #Meta.Environment.Cloud != "local" {
    TemporalServer: "us-east-1.aws.api.temporal.io:7233"
    TemporalNamespace: "totaltalent-qa.bodly"
}