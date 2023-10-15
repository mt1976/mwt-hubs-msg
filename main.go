package main

import (
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/toqueteos/webbrowser"

	xdl "github.com/mt1976/appFrame/dataloader"
	xlg "github.com/mt1976/appFrame/logs"
	xstr "github.com/mt1976/appFrame/strings"
	xtd "github.com/mt1976/appFrame/temp"
)

// Define the global helper functions
var L xlg.XLogger
var T *xdl.Payload
var S *xdl.Payload
var C Common
var D xtd.TempData

type PageData struct {
	AppName        string
	Message        string
	Notes          string
	Image          string
	Refresh        string
	DisplayTime    string
	LastUpdated    string
	StatusList     []Status
	StatusSelected Status
}

type Status struct {
	ID       string
	Text     string
	Image    string
	Selected bool
}

type Common struct {
	AppName        string
	AppVersion     string
	Protocol       string
	URI            string
	Port           string
	QualifiedURI   string
	TemplayDisplay string
	TemplateSetup  string
	Statuses       map[string]string
}

// Define some global vars for the standard fields in the message file
var msgMessage = "message"
var msgNotes = "notes"
var msgImage = "image"
var msgStatus = "status"
var msgUpdated = "updtimestamp"
var msgUpdatedBy = "upduser"
var msgUpdatedOn = "updhost"
var readableTimeFormat = "2006-01-02 15:04:05"

func init() {
	fmt.Println("Initialising")
	L = xlg.New()
	T = xdl.New("translate", "dat", "")
	T.Verbose()
	S = xdl.New("application", "cfg", "")
	S.Verbose()

	fmt.Println("Initialising - Complete")
	C = Common{}
	C.AppName, _ = getDataOrDefault("AppName", "ASDJIODS")
	C.AppVersion, _ = getDataOrDefault("AppVersion", "0.0.0")
	C.Protocol = "http"
	C.URI = "localhost"
	C.Port = "8080"
	C.QualifiedURI = C.Protocol + "://" + C.URI + ":" + C.Port + "/"
	C.TemplayDisplay = "home"
	C.TemplateSetup = "setup"
	si_err := error(nil)
	C.Statuses, si_err = S.GetMap("statusImagesList")
	if si_err != nil {
		L.Fatal(si_err)
	}
	D, err := xtd.Fetch("message")
	if err != nil {
		L.Fatal(err)
	}
	D.Data.Get("message")
	fmt.Println("Initialising - Complete")
	fmt.Println("CONTENT OF S")
	//spew.Dump(S)
	fmt.Println("CONTENT OF D")
	//spew.Dump(D)
	fmt.Println("CONTENT OF C")
	//spew.Dump(C)
	fmt.Println("CONTENT DONE")
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
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/setup", setupHandler)
	mux.HandleFunc("/update", updateHandler)

	// for k, v := range C.Statuses {
	// 	L.WithField("Status", k).WithField("Image", v).Info(T.Get("Status Image"))
	// }

	L.WithField("URI", C.QualifiedURI).Info(T.Get("Listening"))

	webbrowser.Open(C.QualifiedURI)

	L.Fatal(http.ListenAndServe(":"+C.Port, mux))

}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func homeHandler(w http.ResponseWriter, _ *http.Request) {

	//Get Message Data from /temp
	msgData, err := xtd.Fetch("message")
	if err != nil {
		L.Fatal(err)
	}

	message := msgData.Data.Get(msgMessage)
	if message == "" {
		message = "Hello World"
	}

	thisPage := PageData{}
	thisPage.AppName = C.AppName
	thisPage.Message = msgData.Data.Get(msgMessage)
	thisPage.Notes = xstr.MakeStringDisplayable(msgData.Data.Get(msgNotes))
	thisPage.Image = msgData.Data.Get(msgStatus)
	thisPage.Refresh, _ = getRefreshPeriod()
	thisPage.DisplayTime = getNow() + " (" + thisPage.Refresh + ")"

	L.Info("Processing Notes PRE  " + xstr.DQuote(msgData.Data.Get(msgNotes)))
	L.Info("Processing Notes POST " + xstr.MakeStringDisplayable(msgData.Data.Get(msgNotes)))

	tmplName := "html/" + C.TemplayDisplay + ".html"

	t := template.Must(template.ParseFiles(tmplName)) // Create a template.
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, thisPage) // merge.
}

func setupHandler(w http.ResponseWriter, _ *http.Request) {

	//Get Message Data from /temp
	msgData, err := xtd.Fetch("message")
	if err != nil {
		L.Fatal(err)
	}

	message := msgData.Data.Get(msgMessage)
	if message == "" {
		message = "Hello World"
	}

	L.Info("Processing Setup " + xstr.DQuote(msgData.Data.Get(msgNotes)))

	thisPage := PageData{}
	thisPage.AppName = C.AppName
	thisPage.Message = msgData.Data.Get(msgMessage)
	thisPage.Notes = xstr.MakeStringDisplayable(msgData.Data.Get(msgNotes))
	thisPage.Image = msgData.Data.Get(msgImage)
	thisPage.Refresh, _ = getRefreshPeriod()
	thisPage.DisplayTime = getNow() + " (" + thisPage.Refresh + ")"
	thisPage.LastUpdated = getLastUpdated(msgData)
	thisPage.StatusList = getStatusList(msgData)

	tmplName := "html/" + C.TemplateSetup + ".html"

	t := template.Must(template.ParseFiles(tmplName)) // Create a template.
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, thisPage) // merge.
}

func updateHandler(w http.ResponseWriter, r *http.Request) {

	//Get Message Data from /temp
	msgData, err := xtd.Fetch("message")
	if err != nil {
		L.Fatal(err)
	}

	//spew.Dump(msgData)

	L.Info("Processing Update " + r.URL.Path)

	//Update the message
	msgData.Data.Update(msgMessage, r.FormValue(msgMessage))
	msgData.Data.Update(msgNotes, xstr.MakeStringStorable(r.FormValue(msgNotes)))
	msgData.Data.Update(msgImage, "placeholder")

	mood := r.FormValue(msgStatus)
	fmt.Println("Mood: " + mood)
	fmt.Println("Mood: " + mood)
	fmt.Println("Mood: " + mood)
	fmt.Println("Mood: " + mood)
	fmt.Println("Mood: " + mood)

	msgData.Data.Update(msgStatus, mood)

	if mood == "" {
		mood = "happy"
	}

	xtd.Store(msgData)

	http.Redirect(w, r, "/setup", http.StatusFound)
}

// Helper Functions

func getLastUpdated(msgData xtd.TempData) string {
	lastUpdated := msgData.Data.Get("updtimestamp")
	if lastUpdated == "" {
		lastUpdated = "Never"
	}

	//Convert to time
	t, err := time.Parse(xtd.DATETIMEFORMAT, lastUpdated)

	//If error then set to now
	if err != nil {
		return "Unknown"
	}

	//Convert to string & return
	return t.Format(readableTimeFormat)
}

func getStatusList(msgData xtd.TempData) []Status {
	statusList := []Status{}
	for k, v := range C.Statuses {
		status := Status{}
		status.ID = k
		status.Text = k
		status.Image = v
		status.Selected = false
		if msgData.Data.Get("status") == k {
			status.Selected = true
		}
		statusList = append(statusList, status)
	}
	return statusList
}

func getNow() string {
	return time.Now().Format(readableTimeFormat)
}

func getRefreshPeriod() (string, error) {
	return getDataOrDefault("pageRefresh", "30") // Refresh every 30 seconds (by default)
}

func getDataOrDefault(what string, deflt string) (string, error) {

	value, _ := S.GetString(what)
	if value == "" {
		return deflt, nil
	}

	return value, nil
}
