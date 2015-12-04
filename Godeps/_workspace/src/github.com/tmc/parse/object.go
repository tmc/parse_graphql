package parse

// Object is the minimal interface a Parse Object must satisfy.
type Object interface {
	ObjectID() string
}

// ParseObject is a type that satisifies the Object interface and is provided for
// embedding.
type ParseObject struct {
	ID        string `json:"objectId,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// ObjectID returns the ID of the object.
func (o ParseObject) ObjectID() string {
	return o.ID
}

// TODO(tmc): add createdAt/updatedAt to time methods
