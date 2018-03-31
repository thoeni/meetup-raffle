package main

import (
	"reflect"
	"testing"
)

func Test_parseMeetup(t *testing.T) {
	type args struct {
		URL string
	}
	tests := []struct {
		name    string
		args    args
		want    meetup
		wantErr bool
	}{
		{"Parse successfully well formed string", args{"https://www.meetup.com/Go-London-User-Group/events/248895386/"}, meetup{"Go-London-User-Group", "248895386"}, false},
		{"Return error when string is not correct", args{"https://www.meetup.com/Go-London-User-Group/events/"}, meetup{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMeetup(tt.args.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMeetup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMeetup() = %v, want %v", got, tt.want)
			}
		})
	}
}
