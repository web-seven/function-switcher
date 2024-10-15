package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"slices"
	"strings"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	enableAnnotation = "switcher.fn.kndp.io/enabled"
	disableAnnotaion = "switcher.fn.kndp.io/disabled"
)

// Function returns whatever response you ask it to.
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

	switchOn, switchOff, err := f.collectSwitches(req, rsp)
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

// Collect enabled and disable resources names from annotations
func (f *Function) collectSwitches(req *fnv1beta1.RunFunctionRequest, rsp *fnv1beta1.RunFunctionResponse) ([]string, []string, error) {
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
		if strings.Contains(k, enableAnnotation) {
			if switchOn == nil {
				switchOn = []string{}
			}
			v, err = f.renderTemplate(v, enableAnnotation, req)
			if err != nil {
				return []string{}, []string{}, err
			}
			switchOn = append(switchOn, strings.Split(v, ",")...)
		}

		if strings.Contains(k, disableAnnotaion) {
			if switchOff == nil {
				switchOff = []string{}
			}
			v, err = f.renderTemplate(v, disableAnnotaion, req)
			if err != nil {
				return []string{}, []string{}, err
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

// Render Go template against request data
func (f *Function) renderTemplate(tplString string, tplName string, req *fnv1beta1.RunFunctionRequest) (string, error) {
	reqMap, err := convertToMap(req)
	if err != nil {
		return tplString, err
	}

	tmpl, err := template.New(tplName).Parse(tplString)
	if err != nil {
		return tplString, err
	}
	f.log.Debug("constructed request map", "request", reqMap)

	buf := &bytes.Buffer{}

	if err := tmpl.Execute(buf, reqMap); err != nil {
		return tplString, err
	}

	f.log.Debug("rendered manifests", "manifests", buf.String())
	return buf.String(), nil
}

// Convert function request to map
func convertToMap(req *fnv1beta1.RunFunctionRequest) (map[string]any, error) {
	jReq, err := protojson.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal request from proto to json")
	}

	var mReq map[string]any
	if err := json.Unmarshal(jReq, &mReq); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal json to map[string]any")
	}

	return mReq, nil
}
