package parse

// NewQueryIter returns an iterator to iterate over batches of Parse Objects ordered by
// createdAt. It automatically manages Skip values to process the entire set of objects.
//
// The Parse Class to query is inferred from the type of destination.
//
// destination should be a slice of the class you would like to populate.
//
// After every call to Next() destination will have a new batch of objects.
func (c *Client) NewQueryIter(whereClause string, destination interface{}) (*QueryIter, error) {
	className, err := objectTypeNameFromSlice(destination)
	if err != nil {
		return nil, err
	}
	return c.NewQueryClassIter(className, whereClause, destination)
}

// NewQueryClassIter returns an iterator to iterate over batches of Parse Objects ordered by
// createdAt. It automatically manages Skip values to process the entire set of objects.
//
// destination should be a slice of the class you would like to populate.
//
// After every call to Next() destination will have a new batch of objects.
func (c *Client) NewQueryClassIter(className string, whereClause string, destination interface{}) (*QueryIter, error) {
	return &QueryIter{
		client:      c,
		destination: destination,
		where:       whereClause,
		className:   className,
	}, nil
}

// QueryIter allows you to iterate over all objects in a Parse Class.
type QueryIter struct {
	client    *Client
	className string
	where     string

	destination interface{}
	lastErr     error
	processed   int
}

func (i *QueryIter) fetchBatch() error {
	where := QueryOptions{
		Where: i.where,
		Limit: 1000,
		Order: "createdAt",
		Skip:  int(i.processed),
	}
	return i.client.QueryClass(i.className, &where, &i.destination)
}

// Next populates the provided slice with the next batch of objects or it will return false.
func (i *QueryIter) Next() bool {
	if i.lastErr = i.fetchBatch(); i.lastErr != nil {
		return false
	}
	n := len(asInterfaceSlice(i.destination))
	if n == 0 {
		return false
	}
	i.processed += n
	return true
}

// Err returns nil if no errors happened during iteration, or the actual error otherwise.
func (i *QueryIter) Err() error {
	return i.lastErr
}

func asInterfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}
