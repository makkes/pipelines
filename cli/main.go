package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	helmctrlapi "github.com/fluxcd/helm-controller/api/v2beta1"
	ksctrlapi "github.com/fluxcd/kustomize-controller/api/v1beta2"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	PipelineNameLabel  = "pipelines.weave.works/name"
	PipelineOrderLabel = "pipelines.weave.works/stage"
)

func main() {
	cfg := config.GetConfigOrDie()
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		panic(err)
	}

	if err := ksctrlapi.AddToScheme(c.Scheme()); err != nil {
		panic(err)
	}
	if err := helmctrlapi.AddToScheme(c.Scheme()); err != nil {
		panic(err)
	}

	// fetch all Kustomizations that define a pipeline

	pipelines := make(map[string][]ksctrlapi.Kustomization)
	k := ksctrlapi.KustomizationList{}
	if err := c.List(context.Background(), &k, client.HasLabels{
		PipelineNameLabel,
	}); err != nil {
		panic(err)
	}

	for _, ks := range k.Items {
		ks := ks
		pName, ok := ks.GetLabels()[PipelineNameLabel]
		if !ok {
			continue
		}
		pks, ok := pipelines[pName]
		if !ok {
			pks = make([]ksctrlapi.Kustomization, 0)
		}
		pks = append(pks, ks)
		pipelines[pName] = pks
	}

	// sort the Kustomizations by the pipeline stage they're in

	for _, kss := range pipelines {
		sort.SliceStable(kss, func(i, j int) bool {
			ksi := kss[i]
			ksj := kss[j]
			ksiOrder, err := strconv.Atoi(ksi.GetLabels()[PipelineOrderLabel])
			if err != nil {
				return false
			}
			ksjOrder, err := strconv.Atoi(ksj.GetLabels()[PipelineOrderLabel])
			if err != nil {
				return false
			}
			return ksiOrder < ksjOrder
		})
	}

	for pName, pKss := range pipelines {
		fmt.Printf("%s:\n", pName)
		for _, ks := range pKss {
			fmt.Printf("\t%s/%s\n", ks.GetNamespace(), ks.GetName())

			// fetch Deployments in pipeline

			deploys := appsv1.DeploymentList{}
			if err := c.List(context.Background(), &deploys, client.MatchingLabels{
				"kustomize.toolkit.fluxcd.io/name":      ks.Name,
				"kustomize.toolkit.fluxcd.io/namespace": ks.Namespace,
			}); err != nil {
				panic(err)
			}
			for _, deploy := range deploys.Items {
				fmt.Printf("\t\tDeployment/%s/%s: ", deploy.Namespace, deploy.Name)
				for idx, ctr := range deploy.Spec.Template.Spec.Containers {
					fmt.Printf("%s", strings.Split(ctr.Image, ":")[1])
					if idx < len(deploy.Spec.Template.Spec.Containers)-1 {
						fmt.Printf(", ")
					}
				}
				fmt.Println()
			}

			// fetch HelmReleases in pipeline

			hrs := helmctrlapi.HelmReleaseList{}
			if err := c.List(context.Background(), &hrs, client.MatchingLabels{
				"kustomize.toolkit.fluxcd.io/name":      ks.Name,
				"kustomize.toolkit.fluxcd.io/namespace": ks.Namespace,
			}); err != nil {
				panic(err)
			}
			for _, hr := range hrs.Items {
				fmt.Printf("\t\tHelmRelease/%s/%s: %s\n", hr.Namespace, hr.Name, hr.Spec.Chart.Spec.Version)
			}
		}
	}
}
