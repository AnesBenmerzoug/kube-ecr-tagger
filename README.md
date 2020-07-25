
# kube-ecr-tagger

![](https://github.com/AnesBenmerzoug/kube-ecr-tagger/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/AnesBenmerzoug/kube-ecr-tagger/branch/master/graph/badge.svg)](https://codecov.io/gh/AnesBenmerzoug/kube-ecr-tagger)

kube-ecr-tagger is a tool used to complement ECR lifecycles policies by adding a specified tag or tag prefix to all images from ECR that are currently used in your kubernetes cluster.

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
                "ecr:DescribeImages",
                "ecr:BatchGetImage",
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
make install-golangci-lint
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