package lib

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gosimple/slug"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/wader/goutubedl"
	"golang.org/x/image/webp"
	"rul.sh/go-ytmp3/utils"
)

func fetchVideo(video *goutubedl.Result, out string, ch chan error) {
	dl, err := video.Download(context.Background(), "best")
	if err != nil {
		ch <- err
		return
	}
	defer dl.Close()

	f, err := os.Create(out)
	if err != nil {
		ch <- err
		return
	}

	defer f.Close()

	fmt.Println("Downloading...")
	if _, err := io.Copy(f, dl); err != nil {
		ch <- err
		return
	}

	ch <- nil
}

func resizeImage(imgName string, src io.Reader, dst *os.File, width int, height int) error {
	var img image.Image
	var err error

	ext := imgName[strings.LastIndex(imgName, "."):]

	switch ext {
	case "webp":
		img, err = webp.Decode(src)
	default:
		img, _, err = image.Decode(src)
	}

	if err != nil {
		fmt.Println(err)
		return err
	}

	img = imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	return jpeg.Encode(dst, img, nil)
}

func fetchThumbnail(video *goutubedl.Result, thumbnail string, ch chan error) {
	if video.Info.Thumbnail == "" {
		ch <- fmt.Errorf("no thumbnail found")
		return
	}

	fmt.Println("Downloading thumbnail...")

	f, err := os.Create(thumbnail)
	if err != nil {
		ch <- err
		return
	}

	defer f.Close()

	resp, err := http.Get(video.Info.Thumbnail)
	if err != nil {
		ch <- err
		return
	}

	defer resp.Body.Close()

	if err := resizeImage(video.Info.Thumbnail, resp.Body, f, 512, 512); err != nil {
		ch <- err
		return
	}

	ch <- nil
}

type ConvertOptions struct {
	Video     string
	Thumbnail string
	Title     string
	Artist    string
	Album     string
	Output    string
}

func convertToMp3(data ConvertOptions, ch chan error) {
	fmt.Println("Converting...")

	input := []*ffmpeg.Stream{ffmpeg.Input(data.Video).Audio()}
	args := ffmpeg.KwArgs{
		"format":        "mp3",
		"id3v2_version": "3",
		"write_id3v1":   "1",
		"metadata": []string{
			fmt.Sprintf("title=%s", data.Title),
			fmt.Sprintf("artist=%s", data.Artist),
			fmt.Sprintf("album=%s", data.Album),
		},
	}

	if data.Thumbnail != "" {
		input = append(input, ffmpeg.Input(data.Thumbnail).Video())
	}

	if err := ffmpeg.Output(input, data.Output, args).OverWriteOutput().Run(); err != nil {
		ch <- err
		return
	}

	ch <- nil
}

func YtGetVideo(url string) (goutubedl.Result, error) {
	return goutubedl.New(context.Background(), url, goutubedl.Options{})
}

type Yt2Mp3Options struct {
	OutDir    string
	Slug      string
	Thumbnail string
	Title     string
	Artist    string
	Album     string
}

func Yt2Mp3(video *goutubedl.Result, options Yt2Mp3Options) (string, error) {
	if video == nil {
		return "", fmt.Errorf("no video found")
	}

	tmpDir := utils.GetEnv("TMP_DIR", "/tmp")

	title := video.Info.Title
	artist := video.Info.Channel
	album := video.Info.Album

	if video.Info.Artist != "" {
		artist = video.Info.Artist
	}

	videoSlug := options.Slug
	if len(options.Slug) == 0 {
		videoSlug = slug.Make(title)
	}

	videoSrc := fmt.Sprintf("%s/%s.mp4", tmpDir, videoSlug)
	thumbnail := fmt.Sprintf("%s/%s.jpg", tmpDir, videoSlug)
	out := fmt.Sprintf("%s/%s.mp3", options.OutDir, videoSlug)

	if err := os.MkdirAll(options.OutDir, os.ModePerm); err != nil {
		return "", err
	}

	videoCh := make(chan error)
	thumbCh := make(chan error)

	go fetchVideo(video, videoSrc, videoCh)
	go fetchThumbnail(video, thumbnail, thumbCh)

	err := <-videoCh
	if err != nil {
		return "", err
	}

	err = <-thumbCh
	if err != nil {
		thumbnail = ""
	}

	fmt.Println(artist, album)

	convertCh := make(chan error)

	go convertToMp3(ConvertOptions{
		Video:     videoSrc,
		Thumbnail: thumbnail,
		Title:     title,
		Artist:    artist,
		Album:     album,
		Output:    out,
	}, convertCh)

	err = <-convertCh
	if err != nil {
		return "", err
	}

	return out, nil
}
