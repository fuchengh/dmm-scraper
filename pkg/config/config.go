package config

// Output returns the configuration of output
type Output struct {
	Path string
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

type Translate struct {
	Enable       bool
	ApiUrl       string
	ApiKey       string
	Model        string
	Temparature  float64
	TopP         float64
	MaxTokens    int64
	FreqPenalty  float64
	SystemPrompt string
}

// Configs ...
type Configs struct {
	Input     Input
	Output    Output
	Proxy     Proxy
	DMMApi    DMMApi
	Translate Translate
}

// Default ...
func Default() *Configs {
	return &Configs{
		Input: Input{
			Path: ".",
		},
		Output: Output{
			Path: "output/{year}/{num}",
		},
		Proxy: Proxy{
			Enable: false,
			Socket: "",
		},
		DMMApi: DMMApi{
			ApiId:       "",
			AffiliateId: "",
		},
		Translate: Translate{
			Enable:       false,
			ApiUrl:       "",
			ApiKey:       "",
			Model:        "",
			Temparature:  0.0,
			TopP:         0.0,
			MaxTokens:    0,
			FreqPenalty:  0.0,
			SystemPrompt: "",
		},
	}
}
