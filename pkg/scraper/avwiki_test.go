package scraper

import "testing"

func TestAvWikiActors(t *testing.T) {
	BeforeTest()
	tests := []testCase{
		{
			name: "fetchDoc expects no error",
			args: args{
				query: "huntc-202",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fetchAvWikiActors(tt.args.query)
			if len(got) == 0 {
				t.Errorf("fetchAvWikiActors() = %v, want non-empty slice", got)
			} else {
				t.Logf("fetchAvWikiActors() = %v", got)
			}
		})
	}
}