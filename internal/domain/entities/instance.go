package entities

type InstanceConfig struct {
	InstanceID         string            `json:"instance_id"`
	InstanceType       string            `json:"instance_type"`
	AMI                string            `json:"ami"`
	AvailabilityZone   string            `json:"availability_zone"`
	KeyName            string            `json:"key_name"`
	Tags               map[string]string `json:"tags"`
	SecurityGroupIDs   []string          `json:"security_group_ids"`
	SubnetID           string            `json:"subnet_id"`
	IAMInstanceProfile string            `json:"iam_instance_profile"`
	EBSBlockDevices    []EBSBlockDevice  `json:"ebs_block_devices"`
}

type EBSBlockDevice struct {
	DeviceName string `json:"device_name"`
	VolumeSize int    `json:"volume_size"`
	VolumeType string `json:"volume_type"`
}

type DriftReport struct {
	InstanceID string            `json:"instance_id"`
	HasDrift   bool              `json:"has_drift"`
	Changes    map[string]Change `json:"changes"`
}

type Change struct {
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
}
