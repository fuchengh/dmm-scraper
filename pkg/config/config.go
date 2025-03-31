package config

// Output returns the configuration of output
type Output struct {
	Path    string
}

type Input struct {
	Path string
}

// Proxy returns the configuration of proxy
type Proxy struct {
	Enable bool
	Socket string
}

// DMMApi returns the configuration of DMMApi
type DMMApi struct {
	ApiId       string
	AffiliateId string
}

// Configs ...
type Configs struct {
	Input  Input
	Output Output
	Proxy  Proxy
	DMMApi DMMApi
}

// Default ...
func Default() *Configs {
	return &Configs{
		Input: Input{
			Path: ".",
		},
		Output: Output{
			Path:    "output/{year}/{num}",
		},
		Proxy: Proxy{
			Enable: false,
			Socket: "",
		},
		DMMApi: DMMApi{
			ApiId:       "",
			AffiliateId: "",
		},
	}
}
