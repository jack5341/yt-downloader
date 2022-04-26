package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	adress := "localhost:" + port
	fmt.Println("server does work at", adress)

	http.HandleFunc("/watch", stream)
	log.Printf("server does work at %s", adress)
	log.Fatal(http.ListenAndServe(adress, nil))
}

func stream(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query().Get("v")

	if v == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "video id is required argument")
	}

	err := listenAndBroadcast(v, w)

	if err != nil {
		log.Fatal(err)
		fmt.Fprintf(w, "error: %v", err)
		return
	}
}

func listenAndBroadcast(id string, out io.Writer) error {
	url := "www.youtube.com/watch?v=" + id

	r, w := io.Pipe()
	defer r.Close()

	ytdl := exec.Command("youtube-dl", "-o-", url)

	ytdl.Stdout = w
	ytdl.Stderr = os.Stderr

	ffmpeg := exec.Command("ffmpeg", "-i", "/dev/stdin", "-f", "mp3", "-ab", "9600", "-vn", "-")

	ffmpeg.Stdin = r
	ffmpeg.Stdout = out
	ffmpeg.Stderr = os.Stderr

	go func() {
		if err := ytdl.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	err := ffmpeg.Run()
	return err
}
