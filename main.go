package main

import (
	"io"
	"fmt"
	"os"
	"github.com/pkg/errors"
	"encoding/base64"
	"net/http"
	"encoding/json"
	"math/rand"
	"io/ioutil"
	"time"
)

type Attendee struct {
	Member Member `json:"member"`
	Rsvp   Rsvp   `json:"rsvp"`
}

type Member struct {
	Id           float64      `json:"id"`
	Name         string       `json:"name"`
	Photo        Photo        `json:"photo"`
	Role         string       `json:"role"`
	EventContext EventContext `json:"event_context"`
}

type Photo struct {
	ImgURL string `json:"photo_link"`
}

type Rsvp struct {
	Response string `json:"response"`
}

type EventContext struct {
	Host bool `json:"host"`
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Wrong number of arguments. Usage `meetup-raffle 248310043` for meetup with id 248310043")
		os.Exit(2)
	}

	m := os.Args[1]

	attendees, err := getAttendeesForMeetup(m)
	if err != nil {
		fmt.Printf("Could not get attendees: %v", err)
		return
	}

	var a Attendee
	for ; ; {
		a = pickOne(attendees)
		if !a.Member.EventContext.Host &&
			a.Rsvp.Response == "yes" {
			break
		}
	}

	r, err := http.Get(a.Member.Photo.ImgURL)
	if err != nil {
		fmt.Printf("Error while reading URL %v", err)
		return
	}

	fmt.Printf("The winner is...\n\n - [%s]\n\n", a.Member.Name)
	if err := cat(r.Body); err != nil {
		fmt.Printf("could not cat %v", err)
		return
	}
}

func pickOne(attendees []Attendee) Attendee {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	i := r1.Intn(len(attendees))
	a := attendees[i]
	return a
}

func getAttendeesForMeetup(m string) ([]Attendee, error) {

	var attendees []Attendee

	URL := fmt.Sprintf("http://api.meetup.com/Go-London-User-Group/events/%s/attendance", m)
	r, err := http.Get(URL)
	defer r.Body.Close()

	if err != nil {
		return attendees, errors.Wrap(err, fmt.Sprintf("error while reading URL %v, %v", err, r.StatusCode))
	}

	if r.StatusCode != http.StatusOK {
		return attendees, errors.Errorf("meetup id %s not found!", m)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return attendees, errors.Wrap(err, "Error while reading attendees body")
	}

	if err := json.Unmarshal(data, &attendees); err != nil {
		return attendees, errors.Wrap(err, "could not unmarshall attendees")
	}

	return attendees, nil
}

func cat(r io.ReadCloser) error {
	defer r.Close()
	fmt.Printf("\033]1337;File=inline=1:")

	wc := base64.NewEncoder(base64.StdEncoding, os.Stdout)
	_, err := io.Copy(wc, r)
	if err != nil {
		return errors.Wrap(err, "could not encode image")
	}

	if err := wc.Close(); err != nil {
		return errors.Wrap(err, "could not close base64 encoder")
	}

	print("\a\n")
	return nil
}
