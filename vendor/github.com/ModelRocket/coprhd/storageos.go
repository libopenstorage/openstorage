package coprhd

type (
	// ResourceLink represents the uri for a urn
	ResourceLink struct {
		Rel  string `json:"rel,omitempty"`
		Href string `json:"href,omitempty"`
	}

	// ResourceId contains an id string urn for a resource
	ResourceId struct {
		Id string `json:"id"`
	}

	// Resource is a resource id with a link
	Resource struct {
		ResourceId `json:",inline"`
		Link       ResourceLink `json:"link,omitempty"`
	}

	// NameResource is a resource with a name
	NamedResource struct {
		Resource `json:",inline"`
		Name     string `json:"name"`
	}

	// StorageObject is the base object structure for the storageos platform
	StorageObject struct {
		NamedResource `json:",inline"`
		Inactive      bool     `json:"inactive"`
		Global        bool     `json:"global"`
		Remote        bool     `json:"remote"`
		Vdc           Resource `json:"vdc"`
		Tags          []string `json:"tags"`
		Internal      bool     `json:"internal"`
		Project       Resource `json:"project,omitempty"`
		Tenant        Resource `json:"tenant,omitempty"`
		CreationTime  int64    `json:"creation_time"`
		VArray        Resource `json:"varray,omitempty"`
		Owner         string   `json:"owner,omitempty"`
		Type          string   `json:"type,omitempty"`
	}
)
