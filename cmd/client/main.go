package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/gordonklaus/portaudio"
	"golang.org/x/sync/errgroup"
)

func main() {
	userF := flag.String("user", "user", "username")
	listenUserF := flag.String("other-user", "other-user", "username of the other user who you want to call")
	flag.Parse()

	if userF == nil || *userF == "" || listenUserF == nil || *listenUserF == "" {
		panic("user and other-user cannot be empty")
	}

	portaudio.Initialize()
	defer portaudio.Terminate()

	var group errgroup.Group

	group.Go(func() error {
		pr, pw := io.Pipe()

		go func() {
			_, err := http.Post("http://localhost:8080/audio/send?user="+*userF, "audio/wave", pr)
			if err != nil {
				panic(err)
			}
		}()

		err := sendAudio(pw)
		if err != nil {
			println(fmt.Errorf("error sending audio: %w", err).Error())
			return err
		}
		return nil
	})

	group.Go(func() error {
		resp, err := http.Get("http://localhost:8080/audio/receive?user=" + *listenUserF)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		println("start receiving audio")

		err = receiveAudio(resp.Body)
		if err != nil {
			println(fmt.Errorf("error receiving audio: %w", err).Error())
			return err
		}
		return nil
	})

	panic(group.Wait())
}
