/*


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
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// ConditionType represents a Service condition value
const (
	// RoutesReady is set when network configuration has completed.
	RoutesReady apis.ConditionType = "RoutesReady"
	// DefaultPaddleServiceReady is set when default PaddleService has reported readiness.
	DefaultPaddleServiceReady apis.ConditionType = "DefaultPaddleServiceReady"
)

// PaddleService Ready condition is depending on default PaddleService and route readiness condition
// canary readiness condition only present when canary is used and currently does
// not affect PaddleService readiness condition.
var conditionSet = apis.NewLivingConditionSet(
	DefaultPaddleServiceReady,
	RoutesReady,
)

var _ apis.ConditionsAccessor = (*PaddleServiceStatus)(nil)

func (ss *PaddleServiceStatus) InitializeConditions() {
	conditionSet.Manage(ss).InitializeConditions()
}

// IsReady returns if the service is ready to serve the requested configuration.
func (ss *PaddleServiceStatus) IsReady() bool {
	return conditionSet.Manage(ss).IsHappy()
}

// GetCondition returns the condition by name.
func (ss *PaddleServiceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return conditionSet.Manage(ss).GetCondition(t)
}

func (ss *PaddleServiceStatus) PropagateStatus(serviceStatus *knservingv1.ServiceStatus) {
	conditionType := DefaultPaddleServiceReady
	statusSpec := StatusConfigurationSpec{}
	if ss.Default == nil {
		ss.Default = &statusSpec
	}
	if serviceStatus == nil {
		return
	}
	statusSpec.Name = serviceStatus.LatestCreatedRevisionName
	serviceCondition := serviceStatus.GetCondition(knservingv1.ServiceConditionReady)

	switch {
	case serviceCondition == nil:
	case serviceCondition.Status == v1.ConditionUnknown:
		conditionSet.Manage(ss).MarkUnknown(conditionType, serviceCondition.Reason, serviceCondition.Message)
		statusSpec.Hostname = ""
	case serviceCondition.Status == v1.ConditionTrue:
		conditionSet.Manage(ss).MarkTrue(conditionType)
		if serviceStatus.URL != nil {
			statusSpec.Hostname = serviceStatus.URL.Host
		}
	case serviceCondition.Status == v1.ConditionFalse:
		conditionSet.Manage(ss).MarkFalse(conditionType, serviceCondition.Reason, serviceCondition.Message)
		statusSpec.Hostname = ""
	}
	*ss.Default = statusSpec
}
