package maistra

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *ServiceMeshMemberRoll) DeepCopyInto(out *ServiceMeshMemberRoll) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = ServiceMeshMemberRollSpec{
		Members: in.Spec.Members,
	}
}

// DeepCopyObject returns a generically typed copy of an object
func (in *ServiceMeshMemberRoll) DeepCopyObject() runtime.Object {
	out := ServiceMeshMemberRoll{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *ServiceMeshMemberRollList) DeepCopyObject() runtime.Object {
	out := ServiceMeshMemberRollList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]ServiceMeshMemberRoll, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
