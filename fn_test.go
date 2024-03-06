package main

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRunFunction(t *testing.T) {

	type args struct {
		ctx context.Context
		req *fnv1beta1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1beta1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"DesiredResourcesAreRemoved": {
			reason: "Desired resources in annotation field switcher.fn.kndp.io/disabled are removed from the desired state",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Observed: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`
							{
								"apiVersion":"example.org/v1",
								"kind":"XR",
								"metadata": {
									"annotations": {
										"switcher.fn.kndp.io/disabled": "resourceTwo"
									}
								}
							}
							`),
						},
					},
					Desired: &fnv1beta1.State{
						Resources: map[string]*fnv1beta1.Resource{
							"resourceOne": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceTwo": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceThree": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1beta1.State{
						Resources: map[string]*fnv1beta1.Resource{
							"resourceOne": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceThree": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
						},
					},
				},
			},
		},
		"DesiredResourcesAreRetained": {
			reason: "Desired resources in annotation field switcher.fn.kndp.io/enabled are retained in the desired state",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Observed: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`
							{
								"apiVersion":"example.org/v1",
								"kind":"XR",
								"metadata": {
									"annotations": {
										"switcher.fn.kndp.io/enabled": "resourceTwo"
									}
								}
							}
							`),
						},
					},
					Desired: &fnv1beta1.State{
						Resources: map[string]*fnv1beta1.Resource{
							"resourceOne": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceTwo": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceThree": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1beta1.State{
						Resources: map[string]*fnv1beta1.Resource{
							"resourceTwo": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
						},
					},
				},
			},
		},
		"Ensuring preservation of state from previous functions when no annotations are provided": {
			reason: "Desired resources in annotation field switcher.fn.kndp.io/enabled are retained in the desired state",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Observed: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`
							{
								"apiVersion":"example.org/v1",
								"kind":"XR"
							}
							`),
						},
					},
					Desired: &fnv1beta1.State{
						Resources: map[string]*fnv1beta1.Resource{
							"resourceOne": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceTwo": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceThree": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1beta1.State{
						Resources: map[string]*fnv1beta1.Resource{
							"resourceOne": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceTwo": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
							"resourceThree": {
								Resource: resource.MustStructJSON(`
								{
									"apiVersion":"example.org/v1",
									"kind":"Resource"
								}
								`),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)
			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}
