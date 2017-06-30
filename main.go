package main

//go:generate esc -o assets.go assets

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
)

const AppName = "gjvote"
const DbName = AppName + ".db"

// SiteData is stuff that stays the same
type siteData struct {
	Title       string
	Port        int
	SessionName string
	ServerDir   string
	DevMode     bool

	CurrentJam string
}

// pageData is stuff that changes per request
type pageData struct {
	Site           *siteData
	Title          string
	HideTitleImage bool
	SubTitle       string
	Stylesheets    []string
	HeaderScripts  []string
	Scripts        []string
	FlashMessage   string
	FlashClass     string
	LoggedIn       bool
	Menu           []menuItem
	BottomMenu     []menuItem
	HideAdminMenu  bool
	session        *pageSession
	CurrentJam     string
	ClientId       string
	ClientIsAuth   bool
	ClientIsServer bool
	TeamID         string

	PublicMode   int
	TemplateData interface{}
}

type menuItem struct {
	Label    string
	Location string
	Icon     string
}

var sessionSecret = "JCOP5e8ohkTcOzcSMe74"

var sessionStore = sessions.NewCookieStore([]byte(sessionSecret))
var site *siteData
var r *mux.Router

func main() {
	loadConfig()
	dbSaveSiteConfig(site)
	initialize()

	r = mux.NewRouter()
	r.StrictSlash(true)

	if site.DevMode {
		fmt.Println("Operating in Development Mode")
	}
	//s := http.StripPrefix("/assets/", http.FileServer(FS(site.DevMode)))
	//http.Dir(site.ServerDir+"assets/")))
	//r.PathPrefix("/assets/").Handler(s)
	r.PathPrefix("/assets/").Handler(http.FileServer(FS(site.DevMode)))

	// Public Subrouter
	pub := r.PathPrefix("/").Subrouter()
	pub.HandleFunc("/", handleMain)
	pub.HandleFunc("/vote", handlePublicSaveVote)

	// API Subrouter
	//api := r.PathPrefix("/api").Subtrouter()

	// Admin Subrouter
	admin := r.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/", handleAdmin)
	admin.HandleFunc("/dologin", handleAdminDoLogin)
	admin.HandleFunc("/dologout", handleAdminDoLogout)
	admin.HandleFunc("/{category}", handleAdmin)
	admin.HandleFunc("/{category}/{id}", handleAdmin)
	admin.HandleFunc("/{category}/{id}/{function}", handleAdmin)
	admin.HandleFunc("/{category}/{id}/{function}/{subid}", handleAdmin)

	http.Handle("/", r)

	chain := alice.New(loggingHandler).Then(r)

	fmt.Printf("Listening on port %d\n", site.Port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(site.Port), chain))
}

func loadConfig() {
	site = dbGetSiteConfig()

	if len(os.Args) > 1 {
		for _, v := range os.Args {
			key := v
			val := ""
			eqInd := strings.Index(v, "=")
			if eqInd > 0 {
				// It's a key/val argument
				key = v[:eqInd]
				val = v[eqInd+1:]
			}
			switch key {
			case "-title":
				site.Title = val
				fmt.Print("Set site title: ", site.Title, "\n")
			case "-port":
				var tryPort int
				var err error
				if tryPort, err = strconv.Atoi(val); err != nil {
					fmt.Print("Invalid port given: ", val, " (Must be an integer)\n")
					tryPort = site.Port
				}
				// TODO: Make sure a valid port number is given
				site.Port = tryPort
			case "-session-name":
				site.SessionName = val
			case "-server-dir":
				// TODO: Probably check if the given directory is valid
				site.ServerDir = val
			case "-help", "-h", "-?":
				printHelp()
				done()
			case "-dev":
				site.DevMode = true
			case "-reset-defaults":
				resetToDefaults()
				done()
			}
		}
	}
}

func initialize() {
	// Check if the database has been created
	assertError(initDatabase())

	if !dbHasUser() {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Create new Admin user")
		fmt.Print("Email: ")
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)
		var pw1, pw2 []byte
		for string(pw1) != string(pw2) || string(pw1) == "" {
			fmt.Print("Password: ")
			pw1, _ = terminal.ReadPassword(0)
			fmt.Println("")
			fmt.Print("Repeat Password: ")
			pw2, _ = terminal.ReadPassword(0)
			fmt.Println("")
			if string(pw1) != string(pw2) {
				fmt.Println("Entered Passwords don't match!")
			}
		}
		assertError(dbUpdateUserPassword(email, string(pw1)))
	}
	if !dbHasCurrentJam() {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Create New Game Jam")
		fmt.Print("GameJam Name: ")
		gjName, _ := reader.ReadString('\n')
		gjName = strings.TrimSpace(gjName)
		if dbSetCurrentJam(gjName) != nil {
			fmt.Println("Error saving Current Jam")
		}
	}

	jmNm, err := dbGetCurrentJam()
	if err == nil {
		fmt.Println("Current Jam Name: " + jmNm)
	} else {
		fmt.Println(err.Error())
	}
}

func loggingHandler(h http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, h)
}

func InitPageData(w http.ResponseWriter, req *http.Request) *pageData {
	if site.DevMode {
		w.Header().Set("Cache-Control", "no-cache")
	}
	p := new(pageData)
	// Get session
	var err error
	var s *sessions.Session
	if s, err = sessionStore.Get(req, site.SessionName); err != nil {
		http.Error(w, err.Error(), 500)
		return p
	}
	p.session = new(pageSession)
	p.session.session = s
	p.session.req = req
	p.session.w = w

	// First check if we're logged in
	userEmail, _ := p.session.getStringValue("email")
	// With a valid account
	p.LoggedIn = dbIsValidUserEmail(userEmail)

	p.Site = site
	p.SubTitle = "GameJam Voting"
	p.Stylesheets = make([]string, 0, 0)
	p.Stylesheets = append(p.Stylesheets, "/assets/vendor/css/pure-min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/vendor/css/grids-responsive-min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/vendor/font-awesome/css/font-awesome.min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/css/gjvote.css")

	p.HeaderScripts = make([]string, 0, 0)
	p.HeaderScripts = append(p.HeaderScripts, "/assets/vendor/js/snack-min.js")

	p.Scripts = make([]string, 0, 0)
	p.Scripts = append(p.Scripts, "/assets/js/gjvote.js")

	p.FlashMessage, p.FlashClass = p.session.getFlashMessage()
	if p.FlashClass == "" {
		p.FlashClass = "hidden"
	}

	// Build the menu
	if p.LoggedIn {
		p.Menu = append(p.Menu, menuItem{"Admin", "/admin", "fa-key"})
		p.Menu = append(p.Menu, menuItem{"Teams", "/admin/teams", "fa-users"})
		p.Menu = append(p.Menu, menuItem{"Games", "/admin/games", "fa-gamepad"})
		p.Menu = append(p.Menu, menuItem{"Votes", "/admin/votes", "fa-sticky-note"})
		p.Menu = append(p.Menu, menuItem{"Clients", "/admin/clients", "fa-desktop"})

		p.BottomMenu = append(p.BottomMenu, menuItem{"Users", "/admin/users", "fa-user"})
		p.BottomMenu = append(p.BottomMenu, menuItem{"Logout", "/admin/dologout", "fa-sign-out"})
	} else {
		p.BottomMenu = append(p.BottomMenu, menuItem{"Admin", "/admin", "fa-sign-in"})
	}
	p.HideAdminMenu = true

	if p.CurrentJam, err = dbGetCurrentJam(); err != nil {
		p.FlashMessage = "Error Loading Current GameJam: " + err.Error()
		p.FlashClass = "error"
	}

	p.ClientId = p.session.getClientId()
	p.ClientIsAuth = clientIsAuthenticated(p.ClientId, req)
	p.ClientIsServer = clientIsServer(req)
	// TeamID is for team self-administration
	p.TeamID, _ = p.session.getStringValue("teamid")

	// Public Mode
	p.PublicMode = dbGetPublicSiteMode()

	return p
}

func (p *pageData) show(tmplName string, w http.ResponseWriter) error {
	for _, tmpl := range []string{
		"htmlheader.html",
		"admin-menu.html",
		"header.html",
		tmplName,
		"footer.html",
		"htmlfooter.html",
	} {
		if err := outputTemplate(tmpl, p, w); err != nil {
			fmt.Printf("%s\n", err)
			return err
		}
	}
	return nil
}

// outputTemplate
// Spit out a template
func outputTemplate(tmplName string, tmplData interface{}, w http.ResponseWriter) error {
	// TODO: Use embedded files for these... Hopefully?
	_, err := os.Stat("templates/" + tmplName)
	if err == nil {
		t := template.New(tmplName)
		t, _ = t.ParseFiles("templates/" + tmplName)
		return t.Execute(w, tmplData)
	}
	return fmt.Errorf("WebServer: Cannot load template (templates/%s): File not found", tmplName)
}

// redirect can be used only for GET redirects
func redirect(url string, w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, url, 303)
}

func resetToDefaults() {
	def := GetDefaultSiteConfig()
	fmt.Println("Reset settings to defaults?")
	fmt.Print(site.Title, " -> ", def.Title, "\n")
	fmt.Print(site.Port, " -> ", def.Port, "\n")
	fmt.Print(site.SessionName, " -> ", def.SessionName, "\n")
	fmt.Print(site.ServerDir, " -> ", def.ServerDir, "\n")
	fmt.Println("Are you sure? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	conf, _ := reader.ReadString('\n')
	conf = strings.ToUpper(strings.TrimSpace(conf))
	if strings.HasPrefix(conf, "Y") {
		if dbSaveSiteConfig(def) != nil {
			errorExit("Error resetting to defaults")
		}
		fmt.Println("Reset to defaults")
	}
}

func printHelp() {
	help := []string{
		"Game Jam Voting Help",
		"  -help, -h, -?            Print this message",
		"  -dev                     Development mode, load assets from file system",
		"  -port=<port num>         Set the site port",
		"  -session-name=<session>  Set the name of the session to be used",
		"  -server-dir=<directory>  Set the server directory",
		"                           This designates where the database will be saved",
		"                           and where the app will look for files if you're",
		"                           operating in 'development' mode (-dev)",
		"  -title=<title>           Set the site title",
		"  -current-jam=<name>      Change the name of the current jam",
		"  -reset-defaults          Reset all configuration options to defaults",
		"",
	}
	for _, v := range help {
		fmt.Println(v)
	}
}

func done() {
	os.Exit(0)
}

func errorExit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func assertError(err error) {
	if err != nil {
		panic(err)
	}
}
