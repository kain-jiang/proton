/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MongodbOperatorSpec defines the desired state of MongodbOperator
type MongodbOperatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	MongoDBSpec   *MongoDBSpec   `json:"mongodb,omitempty"`
	MgmtSpec      *MgmtSpec      `json:"mgmt,omitempty"`
	ExporterSpec  *ExporterSpec  `json:"exporter,omitempty"`
	LogrotateSpec *LogrotateSpec `json:"logrotate,omitempty"`
	SecretName    string         `json:"secretname,omitempty"`
}

// MongodbOperatorStatus defines the observed state of MongodbOperator
type MongodbOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Replsets map[string]*ReplsetStatus `json:"replsets,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongodbOperator is the Schema for the mongodboperators API
type MongodbOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongodbOperatorSpec   `json:"spec,omitempty"`
	Status MongodbOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongodbOperatorList contains a list of MongodbOperator
type MongodbOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongodbOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongodbOperator{}, &MongodbOperatorList{})
}
