package serverHandler

import (
	"fmt"
	"github.com/kkdai/youtube/v2"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var tpl *template.Template
var videoData *youtube.Video
var client = youtube.Client{}

var data struct {
	Youtubeurl string
	ID         string
	Title      string
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

type myHandler struct {
	mu sync.RWMutex
}

func setContentType(path string) string {
	if strings.HasSuffix(path, ".html") {
		return "text/html"
	} else if strings.HasSuffix(path, ".gohtml") {
		return "text/html"
	} else if strings.HasSuffix(path, ".js") {
		return "application/javascript"
	} else if strings.HasSuffix(path, ".css") {
		return "text/css"
	} //else if strings.HasSuffix(path, ".png") {
	// 	return "image/png"
	// } else if strings.HasSuffix(path, ".svg") {
	// 	return "image/svg+xml"
	// }
	return "text/plain"
}

func (mhd *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	fmt.Printf("l'url est %s\n", path)
	data, err := ioutil.ReadFile(path)

	mhd.mu.Lock()
	defer mhd.mu.Unlock()

	if err == nil {
		var contentType string

		contentType = setContentType(path)

		w.Header().Add("Content-type", contentType)
		w.Write(data)
	} else {
		w.WriteHeader(404)
		w.Write([]byte("404 Page " + http.StatusText(404)))
	}
}

func isServerError(w http.ResponseWriter, err error, page string) {
	if err != nil {
		error := "error while loading index.gohtml"
		fmt.Println(error)
		http.Error(w, error, http.StatusInternalServerError)
	}
}

func goToHTMLPage(page string, w http.ResponseWriter) {

	err := tpl.ExecuteTemplate(w, page, data)

	isServerError(w, err, page)
}

func indexFunction(w http.ResponseWriter, r *http.Request) {

	goToHTMLPage("index.gohtml", w)
}

func loadVideoFunction(w http.ResponseWriter, r *http.Request) {

	var gvErr error
	r.ParseForm()

	for key, value := range r.Form {
		if key == "urlField" && value[0] != "" {
			fmt.Printf("urlField is not empty\n")
			data.Youtubeurl = value[0]
			break
		}
	}

	videoData, gvErr = client.GetVideo(data.Youtubeurl)
	if gvErr != nil {
		fmt.Println("error while loading video from ID")
		os.Exit(0)
	}
	data.Title = videoData.Title
	data.ID = videoData.ID

	fmt.Println("ID: " + data.ID)
	fmt.Println("Title: " + data.Title)

	goToHTMLPage("videoLoadPage.gohtml", w)

}

func downloadVideoFunction(w http.ResponseWriter, r *http.Request) {

	var format string

	r.ParseForm()
	for key, value := range r.Form {
		if key == "format" && value[0] != "" {
			format = value[0]
			break
		}
	}
	if format == "mp4" {
		DownlaodMp4(videoData, w)
	} else if format == "mp3" {
		DownloadMp3()
	}
}

func DownlaodMp4(videoData *youtube.Video, w http.ResponseWriter) {
	var title string
	fmt.Println("je suis le mp4 loader")
	if videoData == nil {
		fmt.Println("videoData is nil. exit!")
		os.Exit(-1)
	}
	fmt.Println("create file......")
	if len(videoData.Title) > 9 {
		title = videoData.Title[0:9]
	} else {
		title = videoData.Title
	}

	go func() {
		file, err := os.Create(title + ".mp4")
		if err != nil {
			fmt.Println("error while create file " + title)
			panic(err)
		}
		defer file.Close()
		fmt.Println("download video......")
		resp, err := client.GetStream(videoData, &videoData.Formats[0])
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			panic(err)
		} else {
			fmt.Println("DOWNLOAD SUCCESS!!!")
		}
	}()
	goToHTMLPage("downloadIsDone.gohtml", w)
}

func DownloadMp3() {
	fmt.Println("je suis le mp3 loader")
}

func New(port string) {

	mux := http.NewServeMux()

	fmt.Printf("listening on %s...\n", port)

	mux.Handle("/", new(myHandler))
	mux.HandleFunc("/index.html", indexFunction)
	mux.HandleFunc("/loadVideo", loadVideoFunction)
	mux.HandleFunc("/download", downloadVideoFunction)

	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatalf("server failed to start: %v\n", err)
	}
}
