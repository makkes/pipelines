# Environment Promotion using `kustomize`

* Applications are defined at a central location in the repository
* Each pipeline is represented in Git by a directory
* Each stage of a pipeline is represented as a directory under `./<NAME>/` where `<NAME>` is the pipeline's name
* Differences between stages are tracked as `kustomize` patches
* Promotion happens by modifying the respective `kustomize` patch file for the specific stage
* Stage 0 is automatically updated using Flux's image update automation

## Getting Started

* Install Flux on the cluster (`flux install --components-extra image-reflector-controller,image-automation-controller`)
* Fork this repo so that you can add your own deploy key used by Flux to update image references
* Connect Flux to the Git repo: `flux create source git pipelines --url=ssh://git@github.com/<YOUR_HANDLE>/pipelines --branch=main --private-key-file=<SSH_DEPLOY_KEY>`
* Create staging Namespaces: `for ns in dev staging prod ; do kubectl create ns "$ns" ; done`
* Bootstrap pipelines Kustomizations: `for ns in dev staging prod ; do k -n "$ns" apply -f base/${ns}/sync.yaml ; done`

After the last step you have three pipelines defined in the cluster, each deploying to a different Namespace. Run the command in this repo to get an overview of them and the apps deployed into each stage:

```sh
cd cli
go run main.go
```

The output will look similar to this:

```
base:
	dev/base-dev
		Deployment/dev/nginx: 1.23.1
		HelmRelease/dev/podinfo: 6.1.6
	staging/base-staging
		Deployment/staging/nginx: 1.22.0
		HelmRelease/staging/podinfo: 5.2.1
	prod/base-prod
		Deployment/prod/nginx: 1.21.6
		HelmRelease/prod/podinfo: 5.2.1
```

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
