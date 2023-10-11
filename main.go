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
	xenv "github.com/mt1976/appFrame/environment"
	xlog "github.com/mt1976/appFrame/logs"
	xsys "github.com/mt1976/appFrame/system"
	xtl "github.com/mt1976/appFrame/translate"
	"github.com/shirou/gopsutil/disk"
)

type PageData struct {
	AppName string
	Apps    []App
	Devices []DeviceInfo
	System  xsys.SystemInfo
}

type App struct {
	Name            string
	DescriptionFull string
	Description     string
	Badge           string
	BadgeContent    string
	Message         string
	Launchers       []Launcher
	Instance        string
	IconPath        string
	Version         string
	Display         string
	IconFileName    string
}

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

func main() {

	xlog.Info(xtl.Get("Starting"))
	//fmt.Println("hello world")
	//test, _ := app.GetEnvironment()

	//xlog.Info("Proteus Hub")

	xenv.Debug()

	spew.Dump(xenv.ApplicationName())
	xlog.Info(xtl.Get("Application Name") + ": " + xenv.ApplicationName())
	xlog.Info(xtl.Get("Application Version") + ": " + xenv.Name())
	xlog.Info(xtl.Get("Application Test") + ": " + xenv.GetOverride("", "transmission-1", "port"))
	// Setup Endpoints
	mux := http.NewServeMux()
	// At least one "mux" handler is required - Dont remove this
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	mux.HandleFunc("/favicon.ico", handlerAssests)

	//mux.HandleFunc("/", TestPage)
	mux.HandleFunc("/", PageDisplay)

	listenON := xenv.Protocol() + "://" + xenv.URI() + ":" + xenv.Port()
	xlog.WithField("URI", listenON).Info(xtl.Get("Listening"))
	//log.Info("Listening...")
	webbrowser.Open(listenON)
	listenPort := xenv.Port()
	xlog.Fatal(http.ListenAndServe(":"+listenPort, mux))

}

func handlerAssests(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func PageDisplay(w http.ResponseWriter, _ *http.Request) {

	//fmt.Fprintf(w, "Hello, there\n")

	xenv.Refresh()

	dhost := xenv.DockerProtocol() + "://" + xenv.DockerURI() + ":" + xenv.DockerPort()

	xlog.WithField(xtl.Get("Host"), dhost).Info(xtl.Get("Connecting"))

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	xlog.WithField(xtl.Get("Host"), dhost).Info(xtl.Get("Connected"))

	thisPage := PageData{}
	thisPage.AppName = xenv.ApplicationName()

	//extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	//parser := parser.NewWithExtensions(extensions)

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
		thisApp.Name = xenv.GetOverride(thisApp.Name, thisApp.Instance, "name")
		//}

		//	thisApp.Name = getOVRvalue(app.ENV.AppPROTOCOL, thisApp.Instance, "name")
		skip := xenv.GetOverride("false", thisApp.Instance, "hide")
		//log.WithFields(log.Fields{"name": thisApp.Name, "instance": thisApp.Instance, "skip": skip}).Info("Skip Test: ")
		if skip == "true" {
			continue
		}
		data := container.Labels["org.opencontainers.image.description"]
		op := serviceHTMLDescriptionGet(xenv.GetExtra(data, thisApp.Instance, "description"))

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
			thisApp.BadgeContent = xtl.Get("Running")
		case "exited":
			thisApp.Badge = "danger"
			thisApp.BadgeContent = xtl.Get("Exited")
		case "created":
			thisApp.Badge = "warning"
			thisApp.BadgeContent = xtl.Get("Created")
		case "paused":
			thisApp.Badge = "info"
			thisApp.BadgeContent = xtl.Get("Paused")
		case "restarting":
			thisApp.Badge = "info"
			thisApp.BadgeContent = xtl.Get("Restarting")
		case "removing":
			thisApp.Badge = "info"
			thisApp.BadgeContent = xtl.Get("Removing")
		case "dead":
			thisApp.Badge = "danger"
			thisApp.BadgeContent = xtl.Get("Dead")
		default:
			thisApp.Badge = "dark"
			thisApp.BadgeContent = xtl.Get("Unknown")
		}

		noPorts := len(container.Ports)
		//log.Info("No of Ports: " + strconv.Itoa(noPorts))
		for i := 0; i < noPorts; i++ {
			//			log.Info("Port: " + strconv.Itoa(int(container.Ports[i].PublicPort)))
			//			log.Info("Int: " + strconv.Itoa(int(container.Ports[i].PrivatePort)))
			nL := Launcher{}
			nL.Name = container.Labels["org.opencontainers.image.title"]
			if nL.Name == "" {
				nL.Name = xenv.GetOverride(xenv.Protocol(), thisApp.Instance, "name")
			}
			pc := xenv.GetOverride(xenv.Protocol(), thisApp.Instance, "protocol")

			nL.AppURI = pc + "://" + xenv.URI() + ":" + strconv.Itoa(int(container.Ports[i].PublicPort)) + "/"
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

	if xenv.AdditionalServices() {
		//log.Info("Adding Additional Services")
		//noSvc := len(app.ENV.AdditionalServicesList)
		//fmt.Printf("noSvc: %v\n", noSvc)
		//fmt.Printf("app.ENV.AdditionalServicesList: %v\n", app.ENV.AdditionalServicesList)
		for _, v := range xenv.AdditionalServicesList() {
			//		log.WithFields(log.Fields{"index": i, "name": v}).Info(app.TextGet("Service"))

			//	thisApp.Name = getOVRvalue(app.ENV.AppPROTOCOL, thisApp.Instance, "name")
			skip := xenv.GetOverride("false", v, "hide")
			//log.WithFields(log.Fields{"name": v, "instance": v, "skip": skip}).Info(app.TextGet("Skip"))
			if skip == "true" {
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

	tmplName := "html/" + xenv.Template() + ".html"

	t := template.Must(template.ParseFiles(tmplName)) // Create a template.
	//	t, _ = t.ParseFiles("html/"+app.ENV.AppTemplate+".html", nil) // Parse template file.
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, thisPage) // merge.
}

func systemInfoGet() xsys.SystemInfo {
	thisSystem := xsys.Get()

	xlog.WithFields(xlog.Fields{xtl.Get("Hostname"): thisSystem.Hostname, xtl.Get("OS"): thisSystem.OS, xtl.Get("Arch"): thisSystem.Arch, xtl.Get("Uptime"): thisSystem.Uptime}).Info(xtl.Get("System"))

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
		xlog.WithFields(xlog.Fields{xtl.Get("mountpoint"): info.Mountpoint, xtl.Get("percent"): info.HumanPercent, xtl.Get("used"): info.HumanUsed, xtl.Get("free"): info.HumanFree, xtl.Get("total"): info.HumanTotal}).Info(xtl.Get("Device"))
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
	test := xenv.GetOverride("", inInstance, "port")
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
	newApp.Name = xenv.GetExtra(inName, inName, "name")
	newApp.Instance = xenv.GetExtra(inName, inName, "instance")
	op := serviceHTMLDescriptionGet(xenv.GetExtra("descr", inName, "description"))
	newApp.Description = op
	newApp.DescriptionFull = op

	newApp.Badge = xenv.GetExtra("info", inName, "badge")
	newApp.BadgeContent = xenv.GetExtra("Unknown", inName, "badgeContent")

	pName := xenv.GetExtra("", inName, "processname")
	if pName != "" {
		pStatus, _ := processStatusGet(pName)
		if pStatus != "" {
			newApp.Badge = "success"
			newApp.BadgeContent = pStatus
		} else {
			newApp.Badge = "dark"
			newApp.BadgeContent = xtl.Get("Not Running")
		}
	}

	newURI := xenv.GetExtra("http", inName, "protocol") + "://" + xenv.GetExtra("127.0.0.1", inName, "ip") + ":" + xenv.GetExtra("", inName, "port") + "/"
	uriPath := xenv.GetExtra("", inName, "path")
	if uriPath != "" {
		newURI = newURI + uriPath
	}

	newApp.Launchers = append(newApp.Launchers, Launcher{Name: newApp.Name, AppURI: newURI, Port: xenv.GetExtra("", inName, "port")})
	newApp.Message = xenv.GetExtra("-", inName, "message")
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
	xlog.WithFields(xlog.Fields{
		xtl.Get("index"):     i,
		xtl.Get("name"):      thisApp.Instance,
		xtl.Get("app"):       thisApp.Name,
		xtl.Get("launchers"): len(thisApp.Launchers),
		xtl.Get("uri"):       aURI,
		xtl.Get("port"):      aPort,
		xtl.Get("Status"):    thisApp.BadgeContent,
		xtl.Get("State"):     thisApp.Message,
		xtl.Get("Version"):   thisApp.Version,
		xtl.Get("Icon"):      thisApp.IconFileName}).Info(xtl.Get("Service"))
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
		return xtl.Get("Unknown"), nil
	} // end if

	return xtl.Get("Running"), nil
}
