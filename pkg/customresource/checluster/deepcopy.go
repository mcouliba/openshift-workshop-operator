package customresource

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *CheCluster) DeepCopyInto(out *CheCluster) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = CheClusterSpec{
		Server:   in.Spec.Server,
		Storage:  in.Spec.Storage,
		Database: in.Spec.Database,
		Auth:     in.Spec.Auth,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *CheCluster) DeepCopyObject() runtime.Object {
	out := CheCluster{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *CheClusterList) DeepCopyObject() runtime.Object {
	out := CheClusterList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]CheCluster, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
