package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	http.HandleFunc("/", linterHandler)
	http.HandleFunc("/upload", uploadFileHandler)

	srv := &http.Server{
		Addr: ":8080",
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)


	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen: %s\n", err)
		}
	}()

	log.Printf("server started, addr: %s\n", srv.Addr)

	<-done
	log.Print("server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %+v\n", err)
	}

	log.Println("server exited properly")
}

func linterHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("error retrieving the file:", err)
		http.Error(w, "Could not process file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	dst, err := os.Create("./thrash/" + handler.Filename)
	if err != nil {
		log.Println("error creating file:", err)
		http.Error(w, "Could not process file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		log.Println("error saving file:", err)
		http.Error(w, "Could not process file", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("file uploaded succesfully"))
}

func processFile()