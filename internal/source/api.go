package source

const (
	RequestPath       = "package/sources/request"
	ResponsePath      = "package/sources/response"
	AssetRequestPath  = "package/asset/request"
	AssetResponsePath = "package/asset/response"
)

type Request struct {
	ID string `json:"id"` // client assigned ID

	// The package path (e.g. "color" or "example.com/foo/bar")
	Package string `json:"package"`
}

type Response struct {
	ID string `json:"id"` // client assigned ID from request

	// The source files in the requested package in the order they should be parsed
	Sources []string `json:"sources"`

	// The asset file names in the requested package
	AssetNames []string `json:"assetNames"`

	// Error if any
	Error string `json:"error,omitempty"`
}

// AssetRequest requests a binary or text asset from a package.
type AssetRequest struct {
	ID       string `json:"id"`
	Package  string `json:"package"`
	Filename string `json:"filename"`
}

// AssetChunkResponse streams a chunk of the requested asset.
type AssetChunkResponse struct {
	ID    string `json:"id"`
	Chunk int    `json:"chunk"`
	Total int    `json:"total"`
	Data  []byte `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
