package main

//go:generate esc -o assets.go assets templates

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
)

// AppName Application name
const AppName = "gjvote"

// The directory that we're storing all of the data in
const DataDir = "data"

// DbName Main Database name, which is <AppName>.db
const DbName = AppName + ".db"

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
	AuthMode       int

	PublicMode   int
	TemplateData interface{}
}

type menuItem struct {
	Label    string
	Location string
	Icon     string
}

var sessionStore *sessions.CookieStore
var r *mux.Router
var m *model

func main() {
	var err error
	if m, err = NewModel(); err != nil {
		errorExit("Unable to initialize Model: " + err.Error())
	}

	loadConfig()
	if err = m.site.SaveToDB(); err != nil {
		errorExit("Unable to save site config to DB: " + err.Error())
	}

	// Save changes to the DB every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			m.saveChanges()
		}
	}()

	initialize()

	// We should have a session secret by now, initialize the store
	sessionStore = sessions.NewCookieStore([]byte(m.site.sessionSecret))

	r = mux.NewRouter()
	r.StrictSlash(true)

	if m.site.DevMode {
		fmt.Println("Operating in Development Mode")
	}
	r.PathPrefix("/assets/").Handler(http.FileServer(FS(m.site.DevMode)))

	// Admin Subrouter
	admin := r.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/", handleAdmin)
	admin.HandleFunc("/dologin", handleAdminDoLogin)
	admin.HandleFunc("/dologout", handleAdminDoLogout)
	admin.HandleFunc("/{category}", handleAdmin)
	admin.HandleFunc("/{category}/{id}", handleAdmin)
	admin.HandleFunc("/{category}/{id}/{function}", handleAdmin)
	admin.HandleFunc("/{category}/{id}/{function}/{subid}", handleAdmin)

	// Public Subrouter
	pub := r.PathPrefix("/").Subrouter()
	pub.HandleFunc("/", handleMain)
	pub.HandleFunc("/{function}", handleMain)
	pub.HandleFunc("/image/{teamid}/{imageid}", handleImageRequest)
	pub.HandleFunc("/thumbnail/{teamid}/{imageid}", handleThumbnailRequest)
	pub.HandleFunc("/team/{id}", handleTeamMgmtRequest)
	pub.HandleFunc("/team/{id}/{function}", handleTeamMgmtRequest)
	pub.HandleFunc("/team/{id}/{function}/{subid}", handleTeamMgmtRequest)

	// API Subrouter
	//api := r.PathPrefix("/api").Subtrouter()
	http.Handle("/", r)

	chain := alice.New(loggingHandler).Then(r)

	// Set up a channel to intercept Ctrl+C for graceful shutdowns
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Save the changes when the app quits
		fmt.Println("\nFinishing up...")
		m.saveChanges()
		os.Exit(0)
	}()

	fmt.Printf("Listening on port %d\n", m.site.Port)
	log.Fatal(http.ListenAndServe(m.site.Ip+":"+strconv.Itoa(m.site.Port), chain))
}

func loadConfig() {
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
				m.site.Title = val
				fmt.Print("Set site title: ", m.site.Title, "\n")
			case "-ip":
				m.site.Ip = val
				fmt.Print("Set site IP: ", m.site.Ip, "\n")
			case "-port":
				var tryPort int
				var err error
				if tryPort, err = strconv.Atoi(val); err != nil {
					fmt.Print("Invalid port given: ", val, " (Must be an integer)\n")
					tryPort = m.site.Port
				}
				// TODO: Make sure a valid port number is given
				m.site.Port = tryPort
			case "-session-name":
				m.site.SessionName = val
			case "-server-dir":
				// TODO: Probably check if the given directory is valid
				m.site.ServerDir = val
			case "-help", "-h", "-?":
				printHelp()
				done()
			case "-dev":
				m.site.DevMode = true
			case "-reset-defaults":
				resetToDefaults()
				done()
			}
		}
	}
}

func initialize() {
	// Test if we have an admin user first
	if !m.hasUser() {
		// Nope, create one
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
		assertError(m.updateUserPassword(email, string(pw1)))
	}

	// Now test if the 'current jam' is named
	if m.jam.Name == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Create New Game Jam")
		fmt.Print("GameJam Name: ")
		gjName, _ := reader.ReadString('\n')
		gjName = strings.TrimSpace(gjName)
		m.jam.Name = gjName
		assertError(m.jam.SaveToDB())
	}

	if m.jam.Name != "" {
		fmt.Println("Current Jam Name: " + m.jam.Name)
	} else {
		fmt.Println("No Jam Name Specified")
	}

	if m.site.sessionSecret == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("A good session secret is like a good password")
		fmt.Print("Create New Session Secret: ")
		sessSc, _ := reader.ReadString('\n')
		sessSc = strings.TrimSpace(sessSc)
		m.site.sessionSecret = sessSc
		assertError(m.site.SaveToDB())
	}
}

func loggingHandler(h http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, h)
}

func InitPageData(w http.ResponseWriter, req *http.Request) *pageData {
	if m.site.DevMode {
		w.Header().Set("Cache-Control", "no-cache")
	}
	p := new(pageData)
	// Get session
	var err error
	var s *sessions.Session
	if s, err = sessionStore.Get(req, m.site.SessionName); err != nil {
		fmt.Println("Session error... Recreating.")
		//http.Error(w, err.Error(), 500)
	}
	p.session = new(pageSession)
	p.session.session = s
	p.session.req = req
	p.session.w = w

	// First check if we're logged in
	userEmail, _ := p.session.getStringValue("email")
	// With a valid account
	p.LoggedIn = m.isValidUserEmail(userEmail)

	p.Site = m.site
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
		p.Menu = append(p.Menu, menuItem{"Archive", "/admin/archive", "fa-archive"})
		p.Menu = append(p.Menu, menuItem{"Clients", "/admin/clients", "fa-desktop"})

		p.BottomMenu = append(p.BottomMenu, menuItem{"Users", "/admin/users", "fa-user"})
		p.BottomMenu = append(p.BottomMenu, menuItem{"Logout", "/admin/dologout", "fa-sign-out"})
	} else {
		p.BottomMenu = append(p.BottomMenu, menuItem{"Admin", "/admin", "fa-sign-in"})
	}
	p.HideAdminMenu = true

	p.ClientId = p.session.getClientId()
	var cl *Client
	if cl, err = m.GetClient(p.ClientId); err != nil {
		// A new client
		cl = NewClient(p.ClientId)
	}
	p.ClientIsAuth = cl.Auth
	p.ClientIsServer = clientIsServer(req)

	// Public Mode
	p.PublicMode = m.site.GetPublicMode()
	// Authentication Mode
	p.AuthMode = m.site.GetAuthMode()

	return p
}

func (p *pageData) show(tmplName string, w http.ResponseWriter) error {
	for _, tmpl := range []string{
		"htmlheader.html",
		"header.html",
		tmplName,
		"footer.html",
		"htmlfooter.html",
	} {
		if err := outputTemplate(tmpl, p, w); err != nil {
			fmt.Printf("Error outputting Template %s\n%s\n", tmpl, err)
			return err
		}
	}
	return nil
}

// outputTemplate
// Spit out a template
func outputTemplate(tmplName string, tmplData interface{}, w http.ResponseWriter) error {
	n := "/templates/" + tmplName
	l := template.Must(template.New("layout").Parse(FSMustString(m.site.DevMode, n)))
	t := template.Must(l.Parse(FSMustString(m.site.DevMode, n)))
	return t.Execute(w, tmplData)
}

// redirect can be used only for GET redirects
func redirect(url string, w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, url, 303)
}

func resetToDefaults() {
	def := NewSiteData(m)
	fmt.Println("Reset settings to defaults?")
	fmt.Print(m.site.Title, " -> ", def.Title, "\n")
	fmt.Print(m.site.Port, " -> ", def.Port, "\n")
	fmt.Print(m.site.SessionName, " -> ", def.SessionName, "\n")
	fmt.Print(m.site.ServerDir, " -> ", def.ServerDir, "\n")
	fmt.Println("Are you sure? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	conf, _ := reader.ReadString('\n')
	conf = strings.ToUpper(strings.TrimSpace(conf))
	if strings.HasPrefix(conf, "Y") {
		if err := def.SaveToDB(); err != nil {
			errorExit("Error resetting to defaults: " + err.Error())
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
