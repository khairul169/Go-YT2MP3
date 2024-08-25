package lib

import "rul.sh/go-ytmp3/utils"

type Task struct {
	Url       string `json:"url"`
	Slug      string `json:"slug"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	IsPending bool   `json:"is_pending"`
	Result    string `json:"result"`
	Error     error  `json:"error"`
}

var tasks = []*Task{}
var queue = []*Task{}

func NewTask(task Task) *Task {
	task.IsPending = true

	queue = append(queue, &task)
	tasks = append(tasks, &task)
	if len(tasks) > 20 {
		tasks = tasks[1:]
	}

	return &task
}

func GetTasks() []*Task {
	return tasks
}

type TaskScheduler struct {
	Ch chan bool
}

func InitTaskScheduler() *TaskScheduler {
	ch := make(chan bool)
	outDir := utils.GetEnv("OUT_DIR", "/tmp")

	go func() {
		for {
			select {
			case <-ch:
				return
			default:
				if len(queue) == 0 {
					continue
				}

				task := queue[0]
				queue = queue[1:]

				video, err := YtGetVideo(task.Url)
				if err != nil {
					task.Error = err
					task.IsPending = false
					continue
				}

				result, err := Yt2Mp3(&video, Yt2Mp3Options{
					OutDir:    outDir,
					Slug:      task.Slug,
					Thumbnail: task.Thumbnail,
					Title:     task.Title,
					Artist:    task.Artist,
					Album:     task.Album,
				})

				task.IsPending = false
				task.Error = err
				task.Result = result
			}
		}
	}()

	return &TaskScheduler{Ch: ch}
}

func (s *TaskScheduler) Stop() {
	s.Ch <- true
}
