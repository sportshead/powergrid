package v10

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Command is a Command resource.
type Command struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec CommandSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CommandList is a collection of Command resources.
type CommandList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	Items []Command `json:"items"`
}

// CommandSpec is the spec of a Command resource.
type CommandSpec struct {
	// ShouldSendDeferred indicates whether to respond with an initial deferred message to Discord. If true, any response from the service will be ignored.
	ShouldSendDeferred bool `json:"shouldSendDeferred,omitempty"`

	ServiceName string `json:"serviceName"`
	// Command represents the Discord command object.
	Command apiextensionsv1.JSON `json:"command"`
}
