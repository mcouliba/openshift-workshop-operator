package nexus

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *Nexus) DeepCopyInto(out *Nexus) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = NexusSpec{
		NexusVolumeSize: in.Spec.NexusVolumeSize,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *Nexus) DeepCopyObject() runtime.Object {
	out := Nexus{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *NexusList) DeepCopyObject() runtime.Object {
	out := NexusList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]Nexus, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
