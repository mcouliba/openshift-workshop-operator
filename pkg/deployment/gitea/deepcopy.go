package customresource

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *Gitea) DeepCopyInto(out *Gitea) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = GiteaSpec{
		GiteaVolumeSize: in.Spec.GiteaVolumeSize,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *Gitea) DeepCopyObject() runtime.Object {
	out := Gitea{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *GiteaList) DeepCopyObject() runtime.Object {
	out := GiteaList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]Gitea, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
