package main

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	var switchOn []string
	var switchOff []string
	meta := v1.ObjectMeta{}
	m := oxr.Resource.Object["metadata"]
	mm, _ := json.Marshal(m)
	json.Unmarshal(mm, &meta)

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

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	for r := range desired {
		if switchOff != nil && slices.Contains[[]string, string](switchOff, string(r)) {
			delete(desired, r)
		}
		if switchOn != nil && !slices.Contains[[]string, string](switchOn, string(r)) {
			delete(desired, r)
		}
	}

	rsp.Desired.Resources = nil

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources in %T", rsp))
		return rsp, nil
	}

	return rsp, nil
}
