package main

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/go-ps"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gomarkdown/markdown"
	"github.com/toqueteos/webbrowser"

	human "github.com/dustin/go-humanize"
	xenv "github.com/mt1976/appFrame/config"
	xdl "github.com/mt1976/appFrame/dataloader"
	xlg "github.com/mt1976/appFrame/logs"
	xsys "github.com/mt1976/appFrame/system"
	"github.com/shirou/gopsutil/disk"
)

// Define the global helper functions
var L xlg.XLogger
var T *xdl.Payload
var S *xdl.Payload
var O *xdl.Payload
var X *xdl.Payload

type PageData struct {
	AppName string
	Apps    []App
	Devices []DeviceInfo
	System  xsys.SystemInfo
}

type App struct {
	Name               string
	DescriptionFull    string
	Description        string
	Badge              string
	BadgeContent       string
	Message            string
	Launchers          []Launcher
	Instance           string
	IconPath           string
	Version            string
	Display            string
	IconFileName       string
	URI                string
	Port               string
	Protocol           string
	QualifiedURI       string
	DockerQualifiedURI string
	DockerURI          string
	DockerPort         string
	DockerProtocol     string
	Template           string
}

var Appl App

type Launcher struct {
	Name   string
	AppURI string
	Port   string
}

type DeviceInfo struct {
	Mountpoint   string
	Percent      float64
	Total        uint64
	Used         uint64
	Free         uint64
	HumanPercent string
	HumanTotal   string
	HumanUsed    string
	HumanFree    string
}

type CPUInfo struct {
	NoCPUs int
	CPUs   string
}

func init() {
	fmt.Println("Initialising - Proteus Hub")
	L = xlg.New()
	T = xdl.New("translate", "dat", "")
	T.Verbose()
	S = xdl.New("system", "env", "")
	S.Verbose()
	fmt.Println("Initialising - Proteus Hub - Complete")
	O = xdl.New("override", "cfg", "")
	O.Verbose()
	X = xdl.New("extra", "cfg", "")
	X.Verbose()
	Appl = App{}
	// Load the Appl Data
	Appl.Name = S.Get("appName")
	Appl.Version = S.Get("appVersion")
	Appl.URI = S.Get("appURI")
	Appl.Port = S.Get("appPort")
	Appl.Protocol = S.Get("appProtocol")
	Appl.QualifiedURI = Appl.Protocol + "://" + Appl.URI + ":" + Appl.Port
	Appl.DockerURI = S.Get("DockerURI")
	Appl.DockerPort = S.Get("DockerPort")
	Appl.DockerProtocol = S.Get("DockerProtocol")
	Appl.DockerQualifiedURI = Appl.DockerProtocol + "://" + Appl.DockerURI + ":" + Appl.DockerPort
	Appl.Template = S.Get("appTemplate")

}

func main() {
	fmt.Println("Starting - Proteus Hub")
	L.Info(T.Get("Starting"))
	//fmt.Println("hello world")
	//test, _ := app.GetEnvironment()

	//L.Info("Proteus Hub")

	xenv.Debug(*S)
	xenv.Debug(*T)
	xenv.Debug(*O)
	xenv.Debug(*X)
	spew.Dump(Appl)

	L.Info(T.Get("Application Name") + ": " + Appl.Name)
	L.Info(T.Get("Application Version") + ": " + Appl.Version)

	// Setup Endpoints
	mux := http.NewServeMux()
	// At least one "mux" handler is required - Dont remove this
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	mux.HandleFunc("/favicon.ico", handlerAssests)

	//mux.HandleFunc("/", TestPage)
	mux.HandleFunc("/", PageDisplay)
	//mux.HandleFunc("/setup", SetUpPage)

	//listenON := Appl.Protocol + "://" + Appl.URI + ":" + Appl.Port + "/"

	L.WithField("URI", Appl.QualifiedURI).Info(T.Get("Listening"))

	webbrowser.Open(Appl.QualifiedURI)

	L.Fatal(http.ListenAndServe(":"+Appl.Port, mux))

}

func handlerAssests(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func PageDisplay(w http.ResponseWriter, _ *http.Request) {

	L.WithField(T.Get("Host"), Appl.DockerQualifiedURI).Info(T.Get("Connecting"))

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	L.WithField(T.Get("Host"), Appl.DockerQualifiedURI).Info(T.Get("Connected"))

	thisPage := PageData{}
	thisPage.AppName = Appl.Name

	//log.Info("Containers List")
	for _, container := range containers {

		thisApp := App{}
		thisApp.Name = html.EscapeString(container.Labels["org.opencontainers.image.title"])
		//log.Info("Processing: " + "MD")
		//log.WithField("data", string(md)).Info("Processing: " + "MD")

		thisApp.Message = container.Status
		thisApp.Instance = container.Names[0][1:]
		//log.Infoln("Processing: /" + thisApp.Instance + "/")
		//if thisApp.Name == "" {
		//thisApp.Name = xenv.GetOverride(thisApp.Name, thisApp.Instance, "name")
		thisApp.Name, err = O.GetStringofCategory(thisApp.Name, "name")
		if err != nil {
			panic(err)
		}
		//}

		//	thisApp.Name = getOVRvalue(app.ENV.AppPROTOCOL, thisApp.Instance, "name")
		//skip := xenv.GetOverride("false", thisApp.Instance, "hide")
		skip, _ := O.GetBoolofCategory(thisApp.Instance, "hide")
		//log.WithFields(log.Fields{"name": thisApp.Name, "instance": thisApp.Instance, "skip": skip}).Info("Skip Test: ")
		if skip {
			continue
		}
		data := container.Labels["org.opencontainers.image.description"]
		//op := serviceHTMLDescriptionGet(X.GetValue(data, thisApp.Instance, "description"))
		desc := X.GetValue(thisApp.Instance, "description")
		if desc == "" {
			desc = data
		}
		op := serviceHTMLDescriptionGet(desc)
		//log.WithField("op", string(op)).Info("Processing: " + "MD")
		//log.Info("Processing: " + "MD END")
		thisApp.DescriptionFull = op
		thisApp.Description = op

		thisApp.IconPath = ""
		thisApp.Version = container.Labels["org.opencontainers.image.version"]
		thisApp.Display, thisApp.IconFileName = serviceFaviconHTMLGet(thisApp)

		switch container.State {
		case "running":
			thisApp.Badge = "success"
			thisApp.BadgeContent = T.Get("Running")
		case "exited":
			thisApp.Badge = "danger"
			thisApp.BadgeContent = T.Get("Exited")
		case "created":
			thisApp.Badge = "warning"
			thisApp.BadgeContent = T.Get("Created")
		case "paused":
			thisApp.Badge = "info"
			thisApp.BadgeContent = T.Get("Paused")
		case "restarting":
			thisApp.Badge = "info"
			thisApp.BadgeContent = T.Get("Restarting")
		case "removing":
			thisApp.Badge = "info"
			thisApp.BadgeContent = T.Get("Removing")
		case "dead":
			thisApp.Badge = "danger"
			thisApp.BadgeContent = T.Get("Dead")
		default:
			thisApp.Badge = "dark"
			thisApp.BadgeContent = T.Get("Unknown")
		}

		noPorts := len(container.Ports)
		//log.Info("No of Ports: " + strconv.Itoa(noPorts))
		for i := 0; i < noPorts; i++ {
			//			log.Info("Port: " + strconv.Itoa(int(container.Ports[i].PublicPort)))
			//			log.Info("Int: " + strconv.Itoa(int(container.Ports[i].PrivatePort)))
			nL := Launcher{}
			nL.Name = container.Labels["org.opencontainers.image.title"]
			if nL.Name == "" {
				nL.Name = X.GetValue(thisApp.Instance, "name")
				//xenv.GetOverride(xenv.Protocol(), thisApp.Instance, "name")
			}
			ptcl := X.GetValue(thisApp.Instance, "protocol")
			pc := Appl.Protocol
			if ptcl != "" {
				pc = ptcl
			}
			nL.AppURI = pc + "://" + Appl.URI + ":" + strconv.Itoa(int(container.Ports[i].PublicPort)) + "/"
			nL.Port = strconv.Itoa(int(container.Ports[i].PublicPort))
			addPort, err := servicePortValidate(thisApp.Instance, nL.Port)
			if err != nil {
				panic(err)
			}
			if addPort {
				thisApp.Launchers = append(thisApp.Launchers, nL)
				//	logApp(thisApp)
			}
		}
		//spew.Dump(thisApp)
		thisPage.Apps = append(thisPage.Apps, thisApp)
	}
	additionalServices, _ := S.GetBool("additionalServices")
	if additionalServices {
		//log.Info("Adding Additional Services")
		//noSvc := len(app.ENV.AdditionalServicesList)
		//fmt.Printf("noSvc: %v\n", noSvc)
		//fmt.Printf("app.ENV.AdditionalServicesList: %v\n", app.ENV.AdditionalServicesList)

		additionalServicesList, _ := S.GetList("additionalServicesList")

		for _, v := range additionalServicesList {
			//		log.WithFields(log.Fields{"index": i, "name": v}).Info(app.TextGet("Service"))

			//	thisApp.Name = getOVRvalue(app.ENV.AppPROTOCOL, thisApp.Instance, "name")

			skip, _ := O.GetBoolofCategory(v, "hide")
			//log.WithFields(log.Fields{"name": v, "instance": v, "skip": skip}).Info(app.TextGet("Skip"))
			if skip {
				continue
			}

			svcDef, _ := serviceInfoGet(v)

			thisPage.Apps = append(thisPage.Apps, svcDef)
			//logApp(svcDef)
		}
	}

	x := len(thisPage.Apps)
	//log.Info("No of Apps: " + strconv.Itoa(x))
	for i := 0; i < x; i++ {
		//	log.Info("SEQ: " + strconv.Itoa(i))
		serviceInfoLog(i, thisPage.Apps[i])
	}

	thisPage.Devices = storageDeviceInfoGet()
	//fmt.Printf("thisPage.Devices: %v\n", thisPage.Devices)

	//	fmt.Printf("thisPage: %v\n", thisPage)
	//	fmt.Fprintln(w, "thisPage: ", thisPage)

	thisPage.System = systemInfoGet()
	//spew.Dump(thisPage.System)

	//fmt.Printf("getSystemInfo(): %v\n", systemInfoGet())

	tmplName := "html/" + Appl.Template + ".html"

	t := template.Must(template.ParseFiles(tmplName)) // Create a template.
	//	t, _ = t.ParseFiles("html/"+app.ENV.AppTemplate+".html", nil) // Parse template file.
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, thisPage) // merge.
}

func systemInfoGet() xsys.SystemInfo {
	thisSystem := xsys.Get()

	L.WithFields(xlg.Fields{T.Get("Hostname"): thisSystem.Hostname, T.Get("OS"): thisSystem.OS, T.Get("Arch"): thisSystem.Arch, T.Get("Uptime"): thisSystem.Uptime}).Info(T.Get("System"))

	//spew.Dump(thisSystem)
	return thisSystem
}

func storageDeviceInfoGet() []DeviceInfo {

	//	fmt.Printf("TestPage\n")

	rtnVal := []DeviceInfo{}

	//formatter := "%-14s %7s %7s %7s %4s %s\n"
	//fmt.Printf(formatter, "Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")

	parts, _ := disk.Partitions(true)
	for _, p := range parts {
		info := DeviceInfo{}
		device := p.Mountpoint
		s, _ := disk.Usage(device)

		if s.Total == 0 {
			continue
		}

		percent := fmt.Sprintf("%2.f%%", s.UsedPercent)

		//

		info.Mountpoint = p.Mountpoint
		info.Percent = s.UsedPercent
		info.Total = s.Total
		info.Used = s.Used
		info.Free = s.Free

		info.HumanPercent = percent
		info.HumanTotal = human.Bytes(s.Total)
		info.HumanUsed = human.Bytes(s.Used)
		info.HumanFree = human.Bytes(s.Free)

		rtnVal = append(rtnVal, info)
		L.WithFields(xlg.Fields{T.Get("mountpoint"): info.Mountpoint, T.Get("percent"): info.HumanPercent, T.Get("used"): info.HumanUsed, T.Get("free"): info.HumanFree, T.Get("total"): info.HumanTotal}).Info(T.Get("Device"))
	}
	return rtnVal
}

func serviceHTMLDescriptionGet(data string) string {
	md := []byte(data)

	op := markdown.ToHTML(md, nil, nil)
	return string(op)
}

func servicePortValidate(inInstance string, inPort string) (bool, error) {
	//log.WithFields(log.Fields{"instance": inInstance, "port": inPort}).Info("Checking Port Validity")
	test := O.GetValue(inInstance, "port")
	//xenv.GetOverride("", inInstance, "port")
	//app.OVR[inInstance+"ports"]
	//fmt.Printf("test: %v\n", test)
	if test == inPort {
		return true, nil
	}
	if test == "" {
		return true, nil
	}
	return false, nil
}

func serviceInfoGet(inName string) (App, error) {

	newApp := App{}
	//x := newFunction(inName,"name")
	newApp.Name = X.GetValue(inName, "name")
	newApp.Instance = X.GetValue(inName, "instance")
	op := serviceHTMLDescriptionGet(X.GetValue(inName, "description"))
	newApp.Description = op
	newApp.DescriptionFull = op

	newApp.Badge = X.GetValue(inName, "badge")
	newApp.BadgeContent = X.GetValue(inName, "badgeContent")

	pName := X.GetValue(inName, "processname")
	if pName != "" {
		pStatus, _ := processStatusGet(pName)
		if pStatus != "" {
			newApp.Badge = "success"
			newApp.BadgeContent = pStatus
		} else {
			newApp.Badge = "dark"
			newApp.BadgeContent = T.Get("Not Running")
		}
	}

	newURI := X.GetValue(inName, "protocol") + "://" + X.GetValue(inName, "ip") + ":" + X.GetValue(inName, "port") + "/"
	uriPath := X.GetValue(inName, "path")
	if uriPath != "" {
		newURI = newURI + uriPath
	}

	newApp.Launchers = append(newApp.Launchers, Launcher{Name: newApp.Name, AppURI: newURI, Port: X.GetValue(inName, "port")})
	newApp.Message = X.GetValue(inName, "message")
	newApp.Display, newApp.IconFileName = serviceFaviconHTMLGet(newApp)
	//newApp.IconFileName
	serviceInfoLog(0, newApp)

	return newApp, nil
}

func serviceFaviconHTMLGet(newApp App) (string, string) {
	defaultIcon := "<i class=\"fas fa-server fa-3x mb-3\"></i>"
	fn := strings.Split(newApp.Instance, "-")
	iconfile := fn[0] + ".png"
	//log.Println("iconfile: " + iconfile)
	//log.Info("Setting Icon : ", newApp.Instance, " ", iconfile)
	// check if file exists in /icons
	if _, err := os.Stat("assets/icons/" + iconfile); err == nil {
		//	log.Info("Icon Found : ", iconfile)
		return "<img src=\"assets/icons/" + iconfile + "\" class=\"img-display fa-2x\" alt=\"" + newApp.Name + "\" />", iconfile
	}
	return defaultIcon, iconfile
}

func serviceInfoLog(i int, thisApp App) {
	//log.Info("LOGGING SEQ: " + strconv.Itoa(i))
	//log.Info("LAUNCERS APP: " + strconv.Itoa(len(thisApp.Launchers)))
	lchr := len(thisApp.Launchers) - 1
	aURI := "N/A"
	aPort := "N/A"
	//log.Info(thisApp.Launchers)

	if lchr >= 0 {
		aURI = thisApp.Launchers[lchr].AppURI
		aPort = thisApp.Launchers[lchr].Port
	}
	L.WithFields(xlg.Fields{
		T.Get("index"):     i,
		T.Get("name"):      thisApp.Instance,
		T.Get("app"):       thisApp.Name,
		T.Get("launchers"): len(thisApp.Launchers),
		T.Get("uri"):       aURI,
		T.Get("port"):      aPort,
		T.Get("Status"):    thisApp.BadgeContent,
		T.Get("State"):     thisApp.Message,
		T.Get("Version"):   thisApp.Version,
		T.Get("Icon"):      thisApp.IconFileName}).Info(T.Get("Service"))
}

func processFindByName(inProcessName string) (int, error) {
	processes, err := ps.Processes()
	if err != nil {
		return 0, err
	} // end if
	for _, process := range processes {
		processName := process.Executable()
		processName = strings.ReplaceAll(processName, "/", "")
		processName = strings.ReplaceAll(processName, ".exe", "")
		//fmt.Println("processName", processName)
		if processName == inProcessName {
			return process.Pid(), nil
		} // end if
	}
	return 0, fmt.Errorf("process %s not found", inProcessName)
}

func processStatusGet(inProcessName string) (string, error) {

	_, err := processFindByName(inProcessName)
	if err != nil {
		return T.Get("Unknown"), nil
	} // end if

	return T.Get("Running"), nil
}
