# ictgj-voting
The ICT GameJam Voting Application

Downloading and Running
----
Download a binary from the list below that is appropriate for your system  
Run the binary, if a database is not found in the current directory, the application will walk you through  
and initial set up (create an admin user, give the site a title, name the current game jam) and a database  
will be created in the current directory  


Command Line Arguments
----

Configuration arguments  
Passing a configuration argument will save the value to the database for future use  
```none
  -title=<title>          Set the title for the site  
  -port=<port>            The port to run the site on  
  -session-name=<name>    A name to use for the session  
  -server-dir=<director>  Directory to use for assets (templates/js/css)  
  -reset-defaults         Reset all of the configurable site settings to their defaults  
                          This only affects the settings that can be set from the command line  
```

Runtime Arguments  
These arguments only affect the current run of the application  
```none
  -help                   Display the application help, breakdown of arguments  
  -dev                    Run in development mode, load assets (templates/js/css) from file system  
                          rather than the binary  
```

Prebuilt Binaries
----
[Linux 64 bit](https://br0xen.com/dowload/ictgj-voting/gjvote.linux64 "Linux 64 bit build")  
[Linux 32 bit](https://br0xen.com/download/ictgj-voting/gjvote.linux386 "Linux 32 bit build")  
[Linux Arm](https://br0xen.com/download/ictgj-voting/gjvote.linuxarm "Linux Arm build")  
[Mac OS](https://br0xen.com/download/ictgj-voting/gjvote.darwin64 "Mac OS build")  
[Windows 64 bit](https://br0xen.com/download/ictgj-voting/gjvote.win64 "Windows 64 bit build")  
[Windows 32 bit](https://br0xen.com/download/ictgj-voting/gjvote.win386 "Windows 32 bit build")  


Building
----
```
go get github.com/devict/ictgj-voting
```


Developing/Contributing Notes
----
Do not make changes to `assets.go`, this file is generated when you run `go generate`  
Pass in the `-dev` flag to enable development mode (load assets from the file system instead of embedded).  
After making changes to assets (templates, javascript, css) be sure to run `go generate` before `go build`  
This regenerates the assets.go file  

Please use the go tooling to match the standard go coding style. For parts that aren't bound by that style,  
either try to match the already existing style, or give a reason why you think it should change.  


Vendorings
----
* 'boltdb' as a data store: https://github.com/boltdb/bolt
* 'boltease' to manipulate the bolt db easier: https://github.com/br0xen/boltease
* Various 'gorilla' libraries for http server stuff: https://github.com/gorilla/
  * context: https://github.com/gorilla/context
  * handlers: https://github.com/gorilla/handlers
  * mux: https://github.com/gorilla/mux
  * securecookie: https://github.com/gorilla/securecookie
  * sessions: https://github.com/gorilla/sessions
* 'alice' for http server middleware: https://github.com/justinas/alice
* 'uuid' for uuid generation:  https://github.com/pborman/uuid
* 'esc' for embedding assets: https://github.com/mjibson/esc


