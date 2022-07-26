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

## Defining a new Pipeline

A pipeline is defined by applying the label `pipelines.weave.works/name` to one or more [Flux Kustomizations](https://fluxcd.io/docs/components/kustomize/kustomization/). Each Kustomization within the same pipeline (i.e. with the same name label) represents one stage of that pipeline. The order of stages is mandated by the label `pipelines.weave.works/stage` holding an integer value >= 0 where a lower value denotes an earlier stage. All applications deployed through one of these Kustomizations are considered part of the given pipeline.

All pipeline Kustomizations can reside in a different Namespace as a pipeline is considered to span a whole cluster or even several clusters.

### Example

In this example we're going to create a pipeline called "podinfo" with the two stages "dev" and "prod". A single application "podinfo" is deployed as part of that pipeline, with version 6.1.6 going to the "dev" stage and 6.0.4 going to the "prod" stage:

```sh
# prepare the "dev" Namespace
$ kubectl create ns podinfo-dev
# create the "dev" source by specifying the 6.1.6 tag to be fetched
$ flux create source git podinfo-dev --url=https://github.com/stefanprodan/podinfo/ --tag=6.1.6
# create the "dev" stage Kustomization
$ flux -n podinfo-dev create ks podinfo --target-namespace=podinfo-dev --source=GitRepository/podinfo-dev.flux-system --path="./kustomize" --prune=true --label=pipelines.weave.works/name=podinfo,pipelines.weave.works/stage=0
# now prepare the "prod" Namespace
$ kubectl create ns podinfo-prod
# create the "prod" source by specifying the 6.0.4 tag this time
$ flux create source git podinfo-prod --url=https://github.com/stefanprodan/podinfo/ --tag=6.0.4
# create the "dev" stage Kustomization
$ flux -n podinfo-prod create ks podinfo --target-namespace=podinfo-prod --source=GitRepository/podinfo-prod.flux-system --path="./kustomize" --prune=true --label=pipelines.weave.works/name=podinfo,pipelines.weave.works/stage=1
```

Now we can use the CLI provided in this repository to introspect the pipeline:

```sh
$ go run main.go
podinfo:
        podinfo-dev/podinfo
                Deployment/podinfo-dev/podinfo: 6.1.6
        podinfo-prod/podinfo
                Deployment/podinfo-prod/podinfo: 6.0.4
```

## Promotion

1. Build and push application image
1. Check that Flux updates the application on dev and the app gets healthy
1. Manually promote the application version from dev to staging by creating a commit changing the `kustomize` patch
1. Check that Flux updates the application on staging and the app gets healthy
1. Manually promote the application version from staging to prod by creating a commit changing the `kustomize` patch
1. Check that Flux updates the application on staging and the app gets healthy
