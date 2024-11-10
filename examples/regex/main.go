/*
   Copyright 2023 The Kubernetes Authors.

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

// Package main is the entrypoint of the %.wasm file, compiled with
// '-target=wasi'. See /guest/RATIONALE.md for details.
package main

import (
	"fmt"
	"regexp"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/api"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/api/proto"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/config"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/klog"
	klogapi "sigs.k8s.io/kube-scheduler-wasm-extension/guest/klog/api"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/plugin"
)

// main is compiled to a WebAssembly function named "_start", called by the
// wasm scheduler plugin during initialization.
func main() {
	p, err := New(klog.Get(), config.Get())
	if err != nil {
		panic(err)
	}
	plugin.Set(p)
}

func New(klog klogapi.Klog, jsonConfig []byte) (api.Plugin, error) {
	return &RegexScheduling{log: klog}, nil
}

// RegexScheduling is a plugin that schedules pods based on a regex annotation.
type RegexScheduling struct {
	log klogapi.Klog
}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "RegexScheduling"
	// RegexAnnotationKey is the key for the pod annotation that defines the regex.
	RegexAnnotationKey = "scheduler.example.com/regex"
)

func (r *RegexScheduling) Filter(state api.CycleState, pod proto.Pod, nodeInfo api.NodeInfo) *api.Status {
	regex, ok := pod.GetAnnotations()[RegexAnnotationKey]
	if !ok {
		return &api.Status{Code: api.StatusCodeSuccess}
	}
	match, err := regexp.MatchString(regex, nodeInfo.Node().GetName())
	if err != nil {
		return &api.Status{Code: api.StatusCodeError, Reason: fmt.Sprintf("Failed to compile regex %q: %s", regex, err)}
	}
	if !match {
		return &api.Status{Code: api.StatusCodeUnschedulable, Reason: fmt.Sprintf("Node %q does not match regex %q", nodeInfo.Node().GetName(), regex)}
	}
	return &api.Status{Code: api.StatusCodeSuccess}
}
