package main

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/toqueteos/webbrowser"

	xdl "github.com/mt1976/appFrame/dataloader"
	xlg "github.com/mt1976/appFrame/logs"
)

// Define the global helper functions
var L xlg.XLogger
var T *xdl.Payload
var S *xdl.Payload
var C Common

type PageData struct {
	AppName string
}

type Common struct {
	AppName         string
	AppVersion      string
	Protocol        string
	URI             string
	Port            string
	QualifiedURI    string
	DisplayTemplate string
	SetupTemplate   string
	StatusImages    map[string]string
}

func init() {
	fmt.Println("Initialising")
	L = xlg.New()
	T = xdl.New("translate", "dat", "")
	T.Verbose()
	S = xdl.New("system", "env", "")
	S.Verbose()
	fmt.Println("Initialising - Complete")
	C = Common{}
	C.AppName, _ = Default("AppName", "ASDJIODS")
	C.AppVersion, _ = Default("AppVersion", "0.0.0")
	C.Protocol = "http"
	C.URI = "localhost"
	C.Port = "8080"
	C.QualifiedURI = C.Protocol + "://" + C.URI + ":" + C.Port + "/"
	C.DisplayTemplate = "home"
	C.SetupTemplate = "setup"
	si_err := error(nil)
	C.StatusImages, si_err = S.GetMap("statusImagesList")
	if si_err != nil {
		L.Fatal(si_err)
	}
}

func main() {
	fmt.Println("Starting - " + C.AppName + " - Version: " + C.AppVersion)
	L.Info(T.Get("Starting"))

	L.Info(T.Get("Application Name") + ": " + C.AppName)
	L.Info(T.Get("Application Version") + ": " + C.AppVersion)

	// Setup Endpoints
	mux := http.NewServeMux()
	// At least one "mux" handler is required - Dont remove this
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	mux.HandleFunc("/favicon.ico", handlerAssests)

	//mux.HandleFunc("/", TestPage)
	mux.HandleFunc("/", PageDisplay)
	//mux.HandleFunc("/setup", SetUpPage)
	for k, v := range C.StatusImages {
		L.WithField("Status", k).WithField("Image", v).Info(T.Get("Status Image"))
	}
	//listenON := Appl.Protocol + "://" + Appl.URI + ":" + Appl.Port + "/"

	L.WithField("URI", C.QualifiedURI).Info(T.Get("Listening"))

	webbrowser.Open(C.QualifiedURI)

	L.Fatal(http.ListenAndServe(":"+C.Port, mux))

}

func handlerAssests(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func PageDisplay(w http.ResponseWriter, _ *http.Request) {

	thisPage := PageData{}
	thisPage.AppName = C.AppName

	tmplName := "html/" + C.DisplayTemplate + ".html"

	t := template.Must(template.ParseFiles(tmplName)) // Create a template.
	//	t, _ = t.ParseFiles("html/"+app.ENV.AppTemplate+".html", nil) // Parse template file.
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, thisPage) // merge.
}

func Default(what string, deflt string) (string, error) {

	value, _ := S.GetString(what)
	if value == "" {
		return deflt, nil
	}

	return value, nil
}
