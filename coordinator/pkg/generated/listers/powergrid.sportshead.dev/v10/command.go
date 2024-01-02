// Code generated by lister-gen. DO NOT EDIT.

package v10

import (
	v10 "github.com/sportshead/powergrid/coordinator/pkg/apis/powergrid.sportshead.dev/v10"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CommandLister helps list Commands.
// All objects returned here must be treated as read-only.
type CommandLister interface {
	// List lists all Commands in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v10.Command, err error)
	// Commands returns an object that can list and get Commands.
	Commands(namespace string) CommandNamespaceLister
	CommandListerExpansion
}

// commandLister implements the CommandLister interface.
type commandLister struct {
	indexer cache.Indexer
}

// NewCommandLister returns a new CommandLister.
func NewCommandLister(indexer cache.Indexer) CommandLister {
	return &commandLister{indexer: indexer}
}

// List lists all Commands in the indexer.
func (s *commandLister) List(selector labels.Selector) (ret []*v10.Command, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v10.Command))
	})
	return ret, err
}

// Commands returns an object that can list and get Commands.
func (s *commandLister) Commands(namespace string) CommandNamespaceLister {
	return commandNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CommandNamespaceLister helps list and get Commands.
// All objects returned here must be treated as read-only.
type CommandNamespaceLister interface {
	// List lists all Commands in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v10.Command, err error)
	// Get retrieves the Command from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v10.Command, error)
	CommandNamespaceListerExpansion
}

// commandNamespaceLister implements the CommandNamespaceLister
// interface.
type commandNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Commands in the indexer for a given namespace.
func (s commandNamespaceLister) List(selector labels.Selector) (ret []*v10.Command, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v10.Command))
	})
	return ret, err
}

// Get retrieves the Command from the indexer for a given namespace and name.
func (s commandNamespaceLister) Get(name string) (*v10.Command, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v10.Resource("command"), name)
	}
	return obj.(*v10.Command), nil
}
