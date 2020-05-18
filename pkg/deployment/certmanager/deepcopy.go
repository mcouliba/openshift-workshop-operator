package certmanager

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *CertManager) DeepCopyInto(out *CertManager) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = CertManagerSpec{}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *CertManager) DeepCopyObject() runtime.Object {
	out := CertManager{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *CertManagerList) DeepCopyObject() runtime.Object {
	out := CertManagerList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]CertManager, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
