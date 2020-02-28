package egoscale

// InstancePoolState represents the state of an instance pool.
type InstancePoolState string

const (
	// InstancePoolCreating creating state.
	InstancePoolCreating InstancePoolState = "creating"
	// InstancePoolRunning running state.
	InstancePoolRunning InstancePoolState = "running"
	// InstancePoolDestroying destroying state.
	InstancePoolDestroying InstancePoolState = "destroying"
	// InstancePoolScalingUp scaling up state.
	InstancePoolScalingUp InstancePoolState = "scaling-up"
	// InstancePoolScalingDown scaling down state.
	InstancePoolScalingDown InstancePoolState = "scaling-down"
)

// InstancePool represents an instance pool.
type InstancePool struct {
	ID                *UUID             `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ServiceOfferingID *UUID             `json:"serviceofferingid"`
	TemplateID        *UUID             `json:"templateid"`
	ZoneID            *UUID             `json:"zoneid"`
	SecurityGroupIDs  []UUID            `json:"securitygroupids"`
	NetworkIDs        []UUID            `json:"networkids"`
	KeyPair           string            `json:"keypair"`
	UserData          string            `json:"userdata"`
	Size              int               `json:"size"`
	RootDiskSize      int               `json:"rootdisksize"`
	State             InstancePoolState `json:"state"`
	VirtualMachines   []VirtualMachine  `json:"virtualmachines"`
}

// CreateInstancePool represents an instance pool creation API request.
type CreateInstancePool struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	ServiceOfferingID *UUID  `json:"serviceofferingid"`
	TemplateID        *UUID  `json:"templateid"`
	ZoneID            *UUID  `json:"zoneid"`
	SecurityGroupIDs  []UUID `json:"securitygroupids,omitempty"`
	NetworkIDs        []UUID `json:"networkids,omitempty"`
	KeyPair           string `json:"keypair,omitempty"`
	UserData          string `json:"userdata,omitempty"`
	Size              int    `json:"size"`
	RootDiskSize      int    `json:"rootdisksize,omitempty"`
	_                 bool   `name:"createInstancePool" description:"Create an Instance Pool"`
}

// CreateInstancePoolResponse represents an instance pool creation API response.
type CreateInstancePoolResponse struct {
	ID                *UUID             `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ServiceOfferingID *UUID             `json:"serviceofferingid"`
	TemplateID        *UUID             `json:"templateid"`
	ZoneID            *UUID             `json:"zoneid"`
	SecurityGroupIDs  []UUID            `json:"securitygroupids"`
	NetworkIDs        []UUID            `json:"networkids"`
	KeyPair           string            `json:"keypair"`
	UserData          string            `json:"userdata"`
	Size              int64             `json:"size"`
	RootDiskSize      int               `json:"rootdisksize"`
	State             InstancePoolState `json:"state"`
}

// Response returns an empty structure to unmarshal an instance pool creation API response into.
func (CreateInstancePool) Response() interface{} {
	return new(CreateInstancePoolResponse)
}

// UpdateInstancePool represents an instance pool update API request.
type UpdateInstancePool struct {
	ID          *UUID  `json:"id"`
	ZoneID      *UUID  `json:"zoneid"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TemplateID  *UUID  `json:"templateid,omitempty"`
	UserData    string `json:"userdata,omitempty"`
	_           bool   `name:"updateInstancePool" description:"Update an Instance Pool"`
}

// UpdateInstancePoolResponse represents an instance pool update API response.
type UpdateInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns an empty structure to unmarshal an instance pool update API response into.
func (UpdateInstancePool) Response() interface{} {
	return new(UpdateInstancePoolResponse)
}

// ScaleInstancePool represents an instance pool scaling API request.
type ScaleInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	Size   int   `json:"size"`
	_      bool  `name:"scaleInstancePool" description:"Scale an Instance Pool"`
}

// ScaleInstancePoolResponse represents an instance pool scaling API response.
type ScaleInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns an empty structure to unmarshal an instance pool scaling API response into.
func (ScaleInstancePool) Response() interface{} {
	return new(ScaleInstancePoolResponse)
}

// DestroyInstancePool represents an instance pool destruction API request.
type DestroyInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"destroyInstancePool" description:"Destroy an Instance Pool"`
}

// DestroyInstancePoolResponse represents an instance pool destruction API response.
type DestroyInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns an empty structure to unmarshal an instance pool destruction API response into.
func (DestroyInstancePool) Response() interface{} {
	return new(DestroyInstancePoolResponse)
}

// GetInstancePool retrieves an instance pool's details.
type GetInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"getInstancePool" description:"Get an Instance Pool"`
}

// GetInstancePoolResponse get instance pool API response.
type GetInstancePoolResponse struct {
	Count         int
	InstancePools []InstancePool `json:"instancepool"`
}

// Response returns an empty structure to unmarshal an instance pool get API response into.
func (GetInstancePool) Response() interface{} {
	return new(GetInstancePoolResponse)
}

// ListInstancePools represents a list instance pool API request.
type ListInstancePools struct {
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"listInstancePools" description:"List Instance Pools"`
}

// ListInstancePoolsResponse represents a list instance pool API response.
type ListInstancePoolsResponse struct {
	Count         int
	InstancePools []InstancePool `json:"instancepool"`
}

// Response returns an empty structure to unmarshal an instance pool list API response into.
func (ListInstancePools) Response() interface{} {
	return new(ListInstancePoolsResponse)
}
