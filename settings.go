package elastigo

// Filter is a filter for aliases
type Filter struct {
	// Term has keys that are the field names
	Term map[string]interface{} `json:"term"`
}

// Alias is an alias for an index
type Alias struct {
	Filter Filter `json:"filter,omitempty"`
}

// Mapping is a single mapping
type Mapping struct {
	// Type is the name of the Elasticsearch type
	Type *string `json:"type,omitempty"`

	// Properties maps the field name to its mapping
	// settings
	Properties map[string]Mapping `json:"properties,omitempty"`

	// Index indicates whether or not we should
	// make this field searchable
	Index *bool `json:"index,omitempty"`

	// Format indicates the format of a particular field
	Format *string `json:"format,omitempty"`

	// Increase search speed for terms aggregations
	EagerGlobalOrdinals *bool `json:"eager_global_ordinals,omitempty"`
}

// Settings contains the number of shards and replicas
type Settings struct {
	NumShards   int32 `json:"number_of_shards,omitempty"`
	NumReplicas int32 `json:"number_of_replicas,omitempty"`
}

// IndexSettings is the settings for a particular index
type IndexSettings struct {
	Aliases  map[string]Alias   `json:"aliases,omitempty"`
	Mappings map[string]Mapping `json:"mappings,omitempty"`
	Settings Settings           `json:"settings,omitempty"`
}

// SetShards sets the number of shards for this index
func (indexSettings *IndexSettings) SetShards(numShards int32) {
	indexSettings.Settings.NumShards = numShards
}

// SetReplicas sets the number of replicas for this index
func (indexSettings *IndexSettings) SetReplicas(numReplicas int32) {
	indexSettings.Settings.NumReplicas = numReplicas
}
