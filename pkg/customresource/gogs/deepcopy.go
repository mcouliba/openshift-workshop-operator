package customresource

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *Gogs) DeepCopyInto(out *Gogs) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = GogsSpec{
		GogsVolumeSize: in.Spec.GogsVolumeSize,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *Gogs) DeepCopyObject() runtime.Object {
	out := Gogs{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *GogsList) DeepCopyObject() runtime.Object {
	out := GogsList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]Gogs, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
