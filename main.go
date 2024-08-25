package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"github.com/joho/godotenv"
	"rul.sh/go-ytmp3/lib"
	"rul.sh/go-ytmp3/types"
	"rul.sh/go-ytmp3/ui"
	"rul.sh/go-ytmp3/utils"
)

func main() {
	godotenv.Load()
	outDir := utils.GetEnv("OUT_DIR", "/tmp")

	app := http.NewServeMux()

	app.HandleFunc("GET /api/info/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if len(url) == 0 {
			http.Error(w, "No video url provided", http.StatusBadRequest)
			return
		}

		video, err := lib.YtGetVideo(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := types.GetVideoInfoRes{
			Url:       url,
			Slug:      slug.Make(video.Info.Title),
			Thumbnail: video.Info.Thumbnail,
			Title:     video.Info.Title,
			Artist:    video.Info.Channel,
			Album:     video.Info.Album,
		}

		if video.Info.Artist != "" {
			data.Artist = video.Info.Artist
		}

		json, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	app.HandleFunc("POST /api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		var data types.CreateTaskBody

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(data.Url) == 0 {
			http.Error(w, "No video url provided", http.StatusBadRequest)
			return
		}

		task := lib.NewTask(lib.Task{
			Url:       data.Url,
			Slug:      data.Slug,
			Thumbnail: data.Thumbnail,
			Title:     data.Title,
			Artist:    data.Artist,
			Album:     data.Album,
		})

		json, err := json.Marshal(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	app.HandleFunc("GET /api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		json, err := json.Marshal(lib.GetTasks())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(json)
	})

	app.HandleFunc("GET /api/get/", func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		isDownload := r.URL.Query().Get("dl") == "true"

		if len(filename) == 0 {
			http.Error(w, "No filename provided", http.StatusBadRequest)
			return
		}

		file, err := os.Open(filepath.Join(outDir, filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		defer file.Close()

		w.Header().Set("Content-Type", "audio/mpeg")
		if isDownload {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		}

		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	ui.ServeUI(app)

	scheduler := lib.InitTaskScheduler()
	defer scheduler.Stop()

	port := utils.GetEnv("PORT", "8080")
	fmt.Printf("Listening on http://localhost:%s\n", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), app); err != nil {
		panic(err)
	}
}
