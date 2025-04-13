package translate

import (
	"dmm-scraper/pkg/config"
	"fmt"
	"testing"
)

func TestTranslate(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				text: "hello, world!",
			},
			want:    "hello, world!",
			wantErr: false,
		},
	}

	// load translate api
	conf, err := config.NewLoader().LoadFile("../../config")
	if err != nil {
		fmt.Println(err)
		conf = config.Default()
	}

	d := New()
	err = d.InitTranslateApi(&conf.Translate)
	if err != nil {
		fmt.Println(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.Translate(tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				fmt.Println(got)
				t.Errorf("Translate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
