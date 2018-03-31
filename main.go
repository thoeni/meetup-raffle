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
	"github.com/briandowns/spinner"
	"regexp"
)

type meetup struct {
	name    string
	eventID string
}

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
		fmt.Printf("Wrong number of arguments. Usage `meetup-raffle https://www.meetup.com/Go-London-User-Group/events/248895386/` \nfor meetup group 'Go-London-User-Group' with event id 248310043")
		os.Exit(2)
	}

	meetupURL := os.Args[1]

	m, err := parseMeetup(meetupURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("The winner is...\n\n")
	spinner := spinner.New(spinner.CharSets[33], 100*time.Millisecond)
	spinner.Start()

	attendees, err := getAttendeesForMeetup(m)
	if err != nil {
		fmt.Printf("\nCould not get attendees: %v", err)
		return
	}

	a := pickOne(attendees)

	r, err := getAttendeeImage(a)
	if err != nil {
		fmt.Printf("Error while picking attendee: %v", err)
		return
	}

	spinner.Stop()

	fmt.Printf(" - [%s]\n\n", a.Member.Name)
	if err := cat(r); err != nil {
		fmt.Printf("could not cat %v", err)
		return
	}
}

func parseMeetup(URL string) (meetup, error) {
	r := regexp.MustCompile("https:\\/\\/www\\.meetup\\.com\\/(.*)\\/events\\/(.*?)\\/")
	m := r.FindAllStringSubmatch(URL, 1)

	if len(m) != 1 {
		return meetup{}, errors.New("Cannot parse meetup URL")
	}

	return meetup{m[0][1], m[0][2]}, nil
}

func getAttendeesForMeetup(m meetup) ([]Attendee, error) {

	var attendees []Attendee

	URL := fmt.Sprintf("http://api.meetup.com/%s/events/%s/attendance", m.name, m.eventID)
	r, err := http.Get(URL)
	defer r.Body.Close()

	if err != nil {
		return attendees, errors.Wrap(err, fmt.Sprintf("error while reading URL [%s] %v, %v", URL, err, r.StatusCode))
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

func pickOne(attendees []Attendee) Attendee {

	var a Attendee

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for ; ; {
		i := r.Intn(len(attendees))
		a = attendees[i]
		if !a.Member.EventContext.Host &&
			a.Rsvp.Response == "yes" {
			break
		}
	}

	return a
}

func getAttendeeImage(a Attendee) (io.ReadCloser, error) {
	var URL string
	URL = a.Member.Photo.ImgURL

	if URL == "" {
		f, err := os.Open("unknown.png")
		if err != nil {
			return nil, errors.Wrap(err, "cannot open local file for unknown emoji")
		}
		return f, nil
	}

	r, err := http.Get(URL)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading URL [%s]", URL)
	}

	return r.Body, nil
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
