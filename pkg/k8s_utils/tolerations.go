/*
Copyright 2017 The Kubernetes Authors.

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
package k8s_utils

/* Code from kubernetes project for merging tolerations */
import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func mergeTolerations(newToleations []corev1.Toleration, destinationTolerations *[]corev1.Toleration) {
	var merged []corev1.Toleration
	all := append(newToleations, *destinationTolerations...)

next:
	for i, t := range all {
		for _, t2 := range merged {
			if isSuperset(t2, t) {
				continue next // t is redundant; ignore it
			}
		}
		if i+1 < len(all) {
			for _, t2 := range all[i+1:] {
				// If the tolerations are equal, prefer the first.
				if !equality.Semantic.DeepEqual(&t, &t2) && isSuperset(t2, t) /* #nosec G601 */ {
					continue next // t is redundant; ignore it
				}
			}
		}
		merged = append(merged, t)
	}
	*destinationTolerations = merged
}

// isSuperset checks whether ss tolerates a superset of t.
func isSuperset(ss, t corev1.Toleration) bool {
	if equality.Semantic.DeepEqual(&t, &ss) {
		return true
	}

	if t.Key != ss.Key &&
		// An empty key with Exists operator means match all keys & values.
		(ss.Key != "" || ss.Operator != corev1.TolerationOpExists) {
		return false
	}

	// An empty effect means match all effects.
	if t.Effect != ss.Effect && ss.Effect != "" {
		return false
	}

	if ss.Effect == corev1.TaintEffectNoExecute {
		if ss.TolerationSeconds != nil {
			if t.TolerationSeconds == nil ||
				*t.TolerationSeconds > *ss.TolerationSeconds {
				return false
			}
		}
	}

	switch ss.Operator {
	case corev1.TolerationOpEqual, "": // empty operator means Equal
		return t.Operator == corev1.TolerationOpEqual && t.Value == ss.Value
	case corev1.TolerationOpExists:
		return true
	default:
		return false
	}
}

/* End code from kubernetes project for merging tolerations */

// compareTolerations checks if all tolerations in the 'first' slice are present in the 'second' slice.
// It returns true if all tolerations in 'first' are found in 'second', false otherwise.
func compareTolerations(first, second []corev1.Toleration) bool {
	for _, t1 := range first {
		found := false
		for _, t2 := range second {
			if reflect.DeepEqual(&t1, &t2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ReconcileTolerations merges the tolerations from the CRD object tolerations into an existing object's
// Tolerations. This is needed because the GKE Autopilot mutating webhook adds tolerations when deployed.
func ReconcileTolerations(newToleations []corev1.Toleration, destinationTolerations *[]corev1.Toleration) (changed bool) {
	changed = !compareTolerations(newToleations, *destinationTolerations)
	mergeTolerations(newToleations, destinationTolerations)
	return changed
}
