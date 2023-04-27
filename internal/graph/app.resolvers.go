package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.29

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
)

// Inbound is the resolver for the inbound field.
func (r *accessPolicyResolver) Inbound(ctx context.Context, obj *model.AccessPolicy) (*model.Inbound, error) {
	ret := &model.Inbound{}
	rules := []*model.Rule{}
	for _, rule := range obj.Inbound.Rules {
		r := model.Rule{}
		r.Application = rule.Application
		r.Namespace = rule.Namespace
		ret.Rules = append(rules, &r)
	}
	return ret, nil
}

// Outbound is the resolver for the outbound field.
func (r *accessPolicyResolver) Outbound(ctx context.Context, obj *model.AccessPolicy) (*model.Outbound, error) {
	ret := &model.Outbound{}
	for _, rule := range obj.Outbound.Rules {
		r := model.Rule{}
		r.Application = rule.Application
		r.Namespace = rule.Namespace
		ret.Rules = append(ret.Rules, &r)
	}

	for _, host := range obj.Outbound.External {
		r := model.External{}
		r.Host = host.Host
		ret.External = append(ret.External, &r)
	}

	return ret, nil
}

// Instances is the resolver for the instances field.
func (r *appResolver) Instances(ctx context.Context, obj *model.App) ([]*model.Instance, error) {
	instances, err := r.K8s.Instances(ctx, obj.GQLVars.Team, obj.Env.Name, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting instances from Kubernetes: %w", err)
	}

	return instances, nil
}

// Resources is the resolver for the resources field.
func (r *appResolver) Resources(ctx context.Context, obj *model.App) (*model.Resources, error) {
	return &model.Resources{
		Limits: &model.Limits{
			CPU:    obj.Resources.Limits.CPU,
			Memory: obj.Resources.Limits.Memory,
		},
		Requests: &model.Requests{
			CPU:    obj.Resources.Requests.CPU,
			Memory: obj.Resources.Requests.Memory,
		},
	}, nil
}

// AccessPolicy returns AccessPolicyResolver implementation.
func (r *Resolver) AccessPolicy() AccessPolicyResolver { return &accessPolicyResolver{r} }

// App returns AppResolver implementation.
func (r *Resolver) App() AppResolver { return &appResolver{r} }

type accessPolicyResolver struct{ *Resolver }
type appResolver struct{ *Resolver }
