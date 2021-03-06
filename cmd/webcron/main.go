package main

import (
	"flag"
	"log"
	"net"
	"os"

	"net/http"

	"github.com/go-chi/chi"

	"github.com/titpetric/go-web-crontab/crontab"
	"github.com/titpetric/factory"
)

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func main() {
	// set up flags
	var (
		configPath = flag.String("config-path", "cron.d/*.cron", "Glob pattern where to load scripts from")
		scriptPath = flag.String("script-path", "cron.scripts/", "Folder from which to run scripts from (working directory)");
		addr = flag.String("addr", ":3000", "Listen address for HTTP server")
	)
	flag.Parse()

	// log to stdout not stderr
	log.SetOutput(os.Stdout)

	// set up database connection
	factory.Database.Add("default", "webcron:webcron@tcp(db1:3306)/webcron?collation=utf8mb4_general_ci")

	db, err := factory.Database.Get()
	handleError(err, "Can't connect to database")
	db.Profiler = &factory.Database.ProfilerStdout

	// listen socket for http server
	log.Println("Starting http server on address " + *addr)
	listener, err := net.Listen("tcp", *addr)
	handleError(err, "Can't listen on addr "+*addr)

	// route options
	routeOptions, err := RouteOptions{}.New()
	handleError(err, "Error creating RouteOptions object")

	// crontab package
	cron, err := crontab.New(db)
	handleError(err, "Error creating Crontab object")
	err = cron.Load(*configPath, *scriptPath)
	handleError(err, "Error loading Crontab configs")

	// mount routes
	r := chi.NewRouter()
	MountRoutes(r, routeOptions, cron)
	http.Serve(listener, r)
}
