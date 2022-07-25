# Environment Promotion using `kustomize`

* Applications are defined at a central location in the repository
* Each pipeline is represented in Git by a directory under `pipelines/`
* Each stage of a pipeline is represented as a directory under `pipelines/<NAME>` where `<NAME>` is the pipeline's name
* Differences between stages are tracked as `kustomize` patches
* Promotion happens by modifying the respective `kustomize` patch file for the specific stage
* Stage 0 is automatically updated using Flux's image update automation

## Getting Started

* Install Flux on the cluster (`flux install --components-extra image-reflector-controller,image-automation-controller`)
* Connect Flux to the Git repo: ``

## Generic DevX

### Promotion

1. Build and push application image
1. Check that Flux updates the application on dev and the app gets healthy
1. Manually promote the application version from dev to staging by creating a commit changing the `kustomize` patch
1. Check that Flux updates the application on staging and the app gets healthy
1. Manually promote the application version from staging to prod by creating a commit changing the `kustomize` patch
1. Check that Flux updates the application on staging and the app gets healthy

### Pipeline Introspection

Each pipeline stage is represented on the cluster by a `Kustomization`. The pipeline name is reflected by the `pipelines.weave.works/name` label on the Kustomization and the order of stages is represented by ascending values of the `pipelines.weave.works/stage` label.
