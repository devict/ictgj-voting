# ictgj-voting
The ICT GameJam Voting Application

## Downloading and Running
1. Download a binary from the list below that is appropriate for your system  
1. Run the binary, if a database is not found in the current directory, the application will walk you through an initial set up:
    1. Create an admin user
    1. Give the site a title
    1. Name the current game jam
    1. A database will be created for you in the current directory  


## Command Line Arguments
### *Configuration arguments*  
Passing a configuration argument will save the value to the database for future use  
```none
  -title=<title>          Set the title for the site  
  -port=<port>            The port to run the site on  
  -session-name=<name>    A name to use for the session  
  -server-dir=<director>  Directory to use for assets (templates/js/css)  
  -reset-defaults         Reset all of the configurable site settings to their defaults  
                          This only affects the settings that can be set from the command line  
```

### *Runtime Arguments*  
These arguments only affect the current run of the application  
```none
  -help                   Display the application help, breakdown of arguments  
  -dev                    Run in development mode, load assets (templates/js/css) from file system  
                          rather than the binary  
```

## Prebuilt Binaries
[Linux 64 bit](https://br0xen.com/dowload/ictgj-voting/ictgj-voting.linux64 "Linux 64 bit build")  
[Linux 32 bit](https://br0xen.com/download/ictgj-voting/ictgj-voting.linux386 "Linux 32 bit build")  
[Linux Arm](https://br0xen.com/download/ictgj-voting/ictgj-voting.linuxarm "Linux Arm build")  
[Mac OS](https://br0xen.com/download/ictgj-voting/ictgj-voting.darwin64 "Mac OS build")  
[Windows 64 bit](https://br0xen.com/download/ictgj-voting/ictgj-voting.win64 "Windows 64 bit build")  
[Windows 32 bit](https://br0xen.com/download/ictgj-voting/ictgj-voting.win386 "Windows 32 bit build")  


## Building
```none
go get github.com/devict/ictgj-voting
```


## Developing/Contributing
### Setup
1. Fork this repo, rather than cloning directly.
1. Then `go get` your github fork, similarly to the command in the Building section above.
1. Make your changes
1. If you changed any template files, run `go generate` to regenerate the `assets.go` file
1. Make and commit your changes, then submit a Pull Request (PR) through GitHub from your fork to `github.com/devict/ictgj-voting`

### Notes
* Pass in the `-dev` flag to enable development mode (load assets from the file system instead of embedded).  
* After making changes to assets (templates, javascript, css) be sure to run `go generate` before `go build` - this regenerates the `assets.go` file  
* Please use the go tooling to match the standard go coding style. 
* For parts that aren't bound by standard go style, either try to match the already existing style, or give a reason why you think it should change.  


## Vendorings
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


