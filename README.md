
# kube-ecr-tagger

![](https://github.com/AnesBenmerzoug/kube-ecr-tagger/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/AnesBenmerzoug/kube-ecr-tagger/branch/master/graph/badge.svg)](https://codecov.io/gh/AnesBenmerzoug/kube-ecr-tagger)

kube-ecr-tagger is a tool used to complement ECR lifecycles policies by adding a specified tag to all images from ECR that are currently used in your kubernetes cluster. 

This is done in order to avoid shooting yourself in the foot by accidentally deleting images that are still being used.

## Requirements

* Working Kubernetes cluster
* IAM Role to tag images on ECR with at least the following policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:DescribeRepositories",
                "ecr:ListImages",
                "ecr:DescribeImages",
                "ecr:BatchGetImage",
                "ecr:ListTagsForResource",
                "ecr:PutImage",
            ],
            "Resource": "*"
        }
    ]
}
```


## Tests

```bash
make test
```

## Linting

Install and run [golanci-lint](https://github.com/golangci/golangci-lint#install)

```bash
# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.22.2
```

```bash
make lint
```

## Build

Dynamically-linked binary:

```bash
make build
```

Statically-linked binary:

```bash
make build-static
```

Docker image:

```bash
make build-image
```