package govultr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ObjectStorageService is the interface to interact with the object storage endpoints on the Vultr API.
// Link: https://www.vultr.com/api/#objectstorage
type ObjectStorageService interface {
	Create(ctx context.Context, objectStoreClusterID int, Label string) (*struct{ ID int `json:"SUBID"` }, error)
	Delete(ctx context.Context, id int) error
	SetLabel(ctx context.Context, id int, label string) error
	List(ctx context.Context, options *ObjectListOptions) ([]ObjectStorage, error)
	Get(ctx context.Context, id int) (*ObjectStorage, error)
	ListCluster(ctx context.Context) ([]ObjectStorageCluster, error)
	RegenerateKeys(ctx context.Context, id int, s3AccessKey string) (*S3Keys, error)
}

// ObjectStorageServiceHandler handles interaction with the firewall rule methods for the Vultr API.
type ObjectStorageServiceHandler struct {
	client *Client
}

// ObjectStorage represents a Vultr Object Storage subscription.
type ObjectStorage struct {
	ID                   int    `json:"SUBID"`
	DateCreated          string `json:"date_created"`
	ObjectStoreClusterID int    `json:"OBJSTORECLUSTERID"`
	RegionID             int    `json:"DCID"`
	Location             string
	Label                string
	Status               string
	S3Keys
}

// ObjectStorageCluster represents a Vultr Object Storage cluster.
type ObjectStorageCluster struct {
	ObjectStoreClusterID int `json:"OBJSTORECLUSTERID"`
	RegionID             int `json:"DCID"`
	Location             string
	Hostname             string
	Deploy               string
}

// S3Keys define your api access to your cluster
type S3Keys struct {
	S3Hostname  string `json:"s3_hostname"`
	S3AccessKey string `json:"s3_access_key"`
	S3SecretKey string `json:"s3_secret_key"`
}

// ObjectListOptions are your optional params you have available to list data.
type ObjectListOptions struct {
	IncludeS3 bool
	Label     string
}

// Create an object storage subscription
func (o *ObjectStorageServiceHandler) Create(ctx context.Context, objectStoreClusterID int, Label string) (*struct{ ID int `json:"SUBID"` }, error) {
	uri := "/v1/objectstorage/create"

	values := url.Values{
		"OBJSTORECLUSTERID": {strconv.Itoa(objectStoreClusterID)},
		"label":             {Label},
	}

	req, err := o.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	id := struct {
		ID int `json:"SUBID"`
	}{}

	err = o.client.DoWithContext(ctx, req, &id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// Delete an object storage subscription.
func (o *ObjectStorageServiceHandler) Delete(ctx context.Context, id int) error {
	uri := "/v1/objectstorage/destroy"

	values := url.Values{
		"SUBID": {strconv.Itoa(id)},
	}

	req, err := o.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = o.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetLabel of an object storage subscription.
func (o *ObjectStorageServiceHandler) SetLabel(ctx context.Context, id int, label string) error {
	uri := "/v1/objectstorage/label_set"

	values := url.Values{
		"SUBID": {strconv.Itoa(id)},
		"label": {label},
	}

	req, err := o.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = o.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List returns all object storage subscriptions on the current account. This includes both pending and active subscriptions.
func (o *ObjectStorageServiceHandler) List(ctx context.Context, options *ObjectListOptions) ([]ObjectStorage, error) {
	uri := "/v1/objectstorage/list"

	req, err := o.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	if options != nil {
		q := req.URL.Query()

		// default behavior is true
		if options.IncludeS3 == false {
			q.Add("include_s3", "false")
		}

		if options.Label != "" {
			q.Add("label", options.Label)
		}

		req.URL.RawQuery = q.Encode()
	}

	var objectStorageMap map[string]ObjectStorage

	err = o.client.DoWithContext(ctx, req, &objectStorageMap)

	if err != nil {
		return nil, err
	}

	var objectStorages []ObjectStorage

	for _, o := range objectStorageMap {
		objectStorages = append(objectStorages, o)
	}

	return objectStorages, nil
}

// Get returns a specified object storage by the provided ID
func (o *ObjectStorageServiceHandler) Get(ctx context.Context, id int) (*ObjectStorage, error) {
	uri := "/v1/objectstorage/list"

	req, err := o.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	if id != 0 {
		q := req.URL.Query()
		q.Add("SUBID", strconv.Itoa(id))
		req.URL.RawQuery = q.Encode()
	}

	objectStorage := new(ObjectStorage)

	err = o.client.DoWithContext(ctx, req, objectStorage)

	if err != nil {
		return nil, err
	}

	return objectStorage, nil
}

// ListCluster returns back your object storage clusters.
// Clusters may be removed over time. The "deploy" field can be used to determine whether or not new deployments are allowed in the cluster.
func (o *ObjectStorageServiceHandler) ListCluster(ctx context.Context) ([]ObjectStorageCluster, error) {
	uri := "/v1/objectstorage/list_cluster"
	req, err := o.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var objectClusterMap map[string]ObjectStorageCluster

	err = o.client.DoWithContext(ctx, req, &objectClusterMap)

	if err != nil {
		return nil, err
	}

	var objectStorageCluster []ObjectStorageCluster

	for _, o := range objectClusterMap {
		objectStorageCluster = append(objectStorageCluster, o)
	}

	return objectStorageCluster, nil
}

// RegenerateKeys of the S3 API Keys for an object storage subscription
func (o *ObjectStorageServiceHandler) RegenerateKeys(ctx context.Context, id int, s3AccessKey string) (*S3Keys, error) {
	uri := "/v1/objectstorage/s3key_regenerate"

	values := url.Values{
		"SUBID":         {strconv.Itoa(id)},
		"s3_access_key": {s3AccessKey},
	}

	req, err := o.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	s3Keys := new(S3Keys)
	err = o.client.DoWithContext(ctx, req, s3Keys)

	if err != nil {
		return nil, err
	}

	return s3Keys, nil
}
