package openapi

// OpenAPISpec represents the OpenAPI 3.1.1 specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers"`
	Paths      map[string]interface{} `json:"paths"`
	Components OpenAPIComponents      `json:"components"`
}

// OpenAPIInfo contains API information
type OpenAPIInfo struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Contact     OpenAPIContact `json:"contact"`
}

// OpenAPIContact contains contact information
type OpenAPIContact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

// OpenAPIServer represents a server
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// OpenAPIComponents contains reusable components
type OpenAPIComponents struct {
	Schemas         map[string]interface{} `json:"schemas"`
	Responses       map[string]interface{} `json:"responses"`
	SecuritySchemes map[string]interface{} `json:"securitySchemes"`
}

// OpenAPIPath represents a path item
type OpenAPIPath struct {
	Summary     string            `json:"summary,omitempty"`
	Description string            `json:"description,omitempty"`
	Get         *OpenAPIOperation `json:"get,omitempty"`
	Post        *OpenAPIOperation `json:"post,omitempty"`
	Put         *OpenAPIOperation `json:"put,omitempty"`
	Delete      *OpenAPIOperation `json:"delete,omitempty"`
	Patch       *OpenAPIOperation `json:"patch,omitempty"`
	Parameters  []interface{}     `json:"parameters,omitempty"`
}

// OpenAPIOperation represents an operation
type OpenAPIOperation struct {
	Summary     string                 `json:"summary"`
	Description string                 `json:"description,omitempty"`
	OperationID string                 `json:"operationId,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Parameters  []interface{}          `json:"parameters,omitempty"`
	RequestBody interface{}            `json:"requestBody,omitempty"`
	Responses   map[string]interface{} `json:"responses"`
	Security    []map[string][]string  `json:"security,omitempty"`
}

// OpenAPIResponse represents a response
type OpenAPIResponse struct {
	Description string                 `json:"description"`
	Content     map[string]interface{} `json:"content,omitempty"`
	Headers     map[string]interface{} `json:"headers,omitempty"`
}

// OpenAPIParameter represents a parameter
type OpenAPIParameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Schema      interface{} `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// OpenAPIRequestBody represents a request body
type OpenAPIRequestBody struct {
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required,omitempty"`
	Content     map[string]interface{} `json:"content"`
}

// OpenAPISchema represents a schema
type OpenAPISchema struct {
	Type                 string                 `json:"type,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Properties           map[string]interface{} `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Items                interface{}            `json:"items,omitempty"`
	AdditionalProperties interface{}            `json:"additionalProperties,omitempty"`
	Example              interface{}            `json:"example,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty"`
	MinItems             *int                   `json:"minItems,omitempty"`
	MaxItems             *int                   `json:"maxItems,omitempty"`
	Pattern              string                 `json:"pattern,omitempty"`
	Ref                  string                 `json:"$ref,omitempty"`
}
