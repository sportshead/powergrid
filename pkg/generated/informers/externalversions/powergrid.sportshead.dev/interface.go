// Code generated by informer-gen. DO NOT EDIT.

package powergrid

import (
	internalinterfaces "github.com/sportshead/powergrid/pkg/generated/informers/externalversions/internalinterfaces"
	v10 "github.com/sportshead/powergrid/pkg/generated/informers/externalversions/powergrid.sportshead.dev/v10"
)

// Interface provides access to each of this group's versions.
type Interface interface {
	// V10 provides access to shared informers for resources in V10.
	V10() v10.Interface
}

type group struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &group{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// V10 returns a new v10.Interface.
func (g *group) V10() v10.Interface {
	return v10.New(g.factory, g.namespace, g.tweakListOptions)
}
