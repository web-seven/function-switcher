package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/pkg/errors"
)

type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {

	rsp := response.To(req, response.DefaultTTL)

	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	s, _ := oxr.Resource.GetString("metadata.annotation.switcher")

	resourcesToSwitch := strings.Split(s, ",")

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	for r := range desired {
		found := false
		for _, rs := range resourcesToSwitch {
			if string(r) == rs {
				found = true
				break
			}
		}

		if found {
			fmt.Println(r)
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
