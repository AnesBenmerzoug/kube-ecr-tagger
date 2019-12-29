
# kube-ecr-tagger

![](https://github.com/AnesBenmerzoug/kube-ecr-tagger/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/AnesBenmerzoug/kube-ecr-tagger/branch/master/graph/badge.svg)](https://codecov.io/gh/AnesBenmerzoug/kube-ecr-tagger)

kube-ecr-tagger is a tool used to complement ECR lifecycles policies by adding a specified tag to all images from ECR that are currently used in your kubernetes cluster. 

This is done in order to avoid shooting yourself in the foot by accidentally deleting images that are still being used.

## Test

```bash
make test
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