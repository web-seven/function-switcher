package main

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	switchOn, switchOff, err := collectSwitches(req, rsp)
	if err != nil {
		return rsp, err
	}

	filterDesired(desired, switchOff, switchOn)

	rsp.Desired.Resources = nil

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources in %T", rsp))
		return rsp, nil
	}

	return rsp, nil
}

func collectSwitches(req *fnv1beta1.RunFunctionRequest, rsp *fnv1beta1.RunFunctionResponse) ([]string, []string, error) {
	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return nil, nil, err
	}
	var switchOn []string
	var switchOff []string
	m := oxr.Resource.Object["metadata"]

	meta, err := toMeta(m)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range meta.Annotations {
		if strings.Contains(k, "switcher.fn.kndp.io/enabled") {
			if switchOn == nil {
				switchOn = []string{}
			}
			switchOn = append(switchOn, strings.Split(v, ",")...)
		}

		if strings.Contains(k, "switcher.fn.kndp.io/disabled") {
			if switchOff == nil {
				switchOff = []string{}
			}
			switchOff = append(switchOff, strings.Split(v, ",")...)
		}
	}

	return switchOn, switchOff, nil
}

func filterDesired(desired map[resource.Name]*resource.DesiredComposed, switchOff []string, switchOn []string) map[resource.Name]*resource.DesiredComposed {

	for r := range desired {
		if switchOff != nil && slices.Contains[[]string, string](switchOff, string(r)) {
			delete(desired, r)
		}
		if switchOn != nil && !slices.Contains[[]string, string](switchOn, string(r)) {
			delete(desired, r)
		}
	}

	return desired
}

func toMeta(m interface{}) (v1.ObjectMeta, error) {
	meta := v1.ObjectMeta{}

	mm, err := json.Marshal(m)
	if err != nil {
		return meta, err
	}
	err = json.Unmarshal(mm, &meta)
	if err != nil {
		return meta, err
	}

	return meta, nil
}
