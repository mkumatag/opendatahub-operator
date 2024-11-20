package modelregistry

import (
	"fmt"

	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	componentsv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/components/v1"
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/datasciencecluster/v1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/pkg/componentsregistry"
	odhdeploy "github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
)

type componentHandler struct{}

func init() { //nolint:gochecknoinits
	cr.Add(&componentHandler{})
}

func (s *componentHandler) GetName() string {
	return componentsv1.ModelRegistryComponentName
}

func (s *componentHandler) GetManagementState(dsc *dscv1.DataScienceCluster) operatorv1.ManagementState {
	return dsc.Spec.Components.ModelRegistry.ManagementState
}

func (s *componentHandler) Init(_ cluster.Platform) error {
	mi := baseManifestInfo(BaseManifestsSourcePath)

	params := make(map[string]string)
	for k, v := range imagesMap {
		params[k] = v
	}
	for k, v := range extraParamsMap {
		params[k] = v
	}

	if err := odhdeploy.ApplyParams(mi.String(), params); err != nil {
		return fmt.Errorf("failed to update images on path %s: %w", mi, err)
	}

	return nil
}

func (s *componentHandler) NewCRObject(dsc *dscv1.DataScienceCluster) client.Object { //nolint:ireturn
	componentAnnotations := make(map[string]string)

	switch dsc.Spec.Components.ModelRegistry.ManagementState {
	case operatorv1.Managed, operatorv1.Removed:
		componentAnnotations[annotations.ManagementStateAnnotation] = string(dsc.Spec.Components.ModelRegistry.ManagementState)
	default:
		// Force and Unmanaged case for unknown values, we do not support these yet
		componentAnnotations[annotations.ManagementStateAnnotation] = "Unknown"
	}

	return client.Object(&componentsv1.ModelRegistry{
		TypeMeta: metav1.TypeMeta{
			Kind:       componentsv1.ModelRegistryKind,
			APIVersion: componentsv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        componentsv1.ModelRegistryInstanceName,
			Annotations: componentAnnotations,
		},
		Spec: componentsv1.ModelRegistrySpec{
			ModelRegistryCommonSpec: dsc.Spec.Components.ModelRegistry.ModelRegistryCommonSpec,
		},
	})
}
