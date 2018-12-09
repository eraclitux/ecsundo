# ecsundo

[![Go Report Card](https://goreportcard.com/badge/github.com/eraclitux/ecsundo)](https://goreportcard.com/report/github.com/eraclitux/ecsundo)
[![CircleCI](https://circleci.com/gh/eraclitux/ecsundo.svg?style=svg)](https://circleci.com/gh/eraclitux/ecsundo)

`ecsundo` is cli tool able to rollback ECS
[services](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html)
in a given cluster to their previous or to specific
[task](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definitions.html)
versions.

It is capable of making and restoring "snapshots" of service versions.

## Motivations

Probably your code is automatically deployed by a CI/CD pipeline.
If something goes wrong it may take several minutes to bring a rollback commit in production, with `ecsundo` this time is cut to seconds.

## Prerequisites

To use this tool, the deploy procedure must involve creating a new task version with a new image tag.
A great way to tag container images is to include the commit hash of the code that they contain, doing so it will
be possible to find the task version you want to come back to if problems arise.
[Learn more](https://medium.com/@eraclitux/deployment-rollback-in-a-containers-world-aws-ecs-edition-4bc8e34c0d5a) about this.

## Usage examples

Rollback _all_ services in a cluster to previous version:

```
$ ecsundo cluster <cluster-name>
```

Rollback a service to the previous version:

```
$ ecsundo service -c <cluster-name> <service-name>
```

Make a _snapshot_ of all services versions in a cluster:

```
$ ecsundo cluster snapshot <cluster-name>
```

Restore a _snapshot_ of a versions of all services in a cluster:

```
$ ecsundo cluster restore <cluster-name>
```

To learn more, use on line help:

```
$ ecsundo help
```

## Configuration

To avoid to specify cluster name, an environment variable can be used:

```
ECSUNDO_CLUSTER=<cluster-name>
```

or a configuration file (default path `~/.ecsundo.yml`):

```
cluster: <cluster-name>
```

Proper **permissions** must be granted for the tool to operate properly.
If you install this tool inside AWS, the best way, from a security standpoint, is to use an IAM role that lets you avoid copying around `AWS_SECRETS`. The role should have at least this permissions:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "ecs:ListAttributes",
                "ecs:DescribeTaskDefinition",
                "ecs:DescribeClusters",
                "ecs:ListServices",
                "ecs:UpdateService",
                "ecs:ListTasks",
                "ecs:ListTaskDefinitionFamilies",
                "ecs:RegisterTaskDefinition",
                "ecs:DescribeServices",
                "ecs:ListContainerInstances",
                "ecs:DescribeContainerInstances",
                "ecs:DescribeTasks",
                "ecs:ListTaskDefinitions",
                "ecs:ListClusters"
            ],
            "Resource": "*"
        }
    ]
}
```

### Use from local shell

A prerequisite is that aws-cli is installed and configured, this is true if credentials file exists (e.g. `ls ~/.aws/credentials`). To use these credentials:

```
AWS_SDK_LOAD_CONFIG=1 ecsundo cluster snapshot <my-cluster>
```

Your credentials must have at least same permissions of the role above.

## Installation

[//]: # "Precompiled binaries can be found [here](https://github.com/eraclitux/ecsundo/releases)."

To install the latest (unstable) version:

```
go get -u github.com/eraclitux/ecsundo
```
