package apis

import (
	"github.com/CSYE-7374-Advanced-Cloud-Computing/operator/pkg/apis/app/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
