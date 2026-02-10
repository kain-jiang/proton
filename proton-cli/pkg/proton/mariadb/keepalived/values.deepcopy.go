package keepalived

func (in *HelmValues) DeepCopyInto(out *HelmValues) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(HelmValuesImage)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.RBAC != nil {
		in, out := &in.RBAC, &out.RBAC
		*out = new(HelmValuesRBAC)
		(*in).DeepCopyInto(*out)
	}
	if in.VIP != nil {
		in, out := &in.VIP, &out.VIP
		*out = new(HelmValuesVIP)
		(*in).DeepCopyInto(*out)
	}
}

func (in *HelmValues) DeepCopy() *HelmValues {
	if in == nil {
		return nil
	}
	out := new(HelmValues)
	in.DeepCopyInto(out)
	return out
}

func (in *HelmValuesImage) DeepCopyInto(out *HelmValuesImage) {
	*out = *in
}
func (in *HelmValuesRBAC) DeepCopyInto(out *HelmValuesRBAC) {
	*out = *in
}
func (in *HelmValuesVIP) DeepCopyInto(out *HelmValuesVIP) {
	*out = *in
}
