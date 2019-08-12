package maistra

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *ServiceMeshControlPlane) DeepCopyInto(out *ServiceMeshControlPlane) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = ServiceMeshControlPlaneSpec{
		Istio: in.Spec.Istio,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *ServiceMeshControlPlane) DeepCopyObject() runtime.Object {
	out := ServiceMeshControlPlane{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *ServiceMeshControlPlaneList) DeepCopyObject() runtime.Object {
	out := ServiceMeshControlPlaneList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]ServiceMeshControlPlane, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
