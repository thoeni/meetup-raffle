package main

import (
	"reflect"
	"testing"
	"net/http/httptest"
	"net/http"
	"fmt"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
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

func Test_getAttendeesForMeetup(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		b, err := ioutil.ReadFile("testdata/meetup-attendance.json")
		if err != nil {
			fmt.Println("Can't open test data", err)
			t.FailNow()
		}
		fmt.Fprintln(w, string(b))
	}))
	defer ts.Close()

	mc := MeetupClient{ts.URL}

	a, err := mc.getAttendeesForMeetup(meetup{"test", "1234"})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(a))
	assert.Equal(t, "John D.", a[0].Member.Name)
	assert.Equal(t, "Mary J.", a[1].Member.Name)
}
