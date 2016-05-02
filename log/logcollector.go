package main

import (
	"flag"
	"fmt"
	"net/http"
	"log"
	"strconv"
	"errors"
	"github.com/deep-patel/log-collection-server/utils"
)

// Job holds the attributes needed to perform unit of work.
type Job struct {
	LogStr  string
	ClientIP string
}

// NewWorker creates takes a numeric id and a channel w/ worker pool.
func NewWorker(id int, workerPool chan chan Job) Worker {
	return Worker{
		id:         id,
		jobQueue:   make(chan Job),
		workerPool: workerPool,
		quitChan:   make(chan bool),
	}
}

type Worker struct {
	id         int
	jobQueue   chan Job
	workerPool chan chan Job
	quitChan   chan bool
}

func (w Worker) start() {
	go func() {
		for {
			// Add my jobQueue to the worker pool.
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				utils.WriteLog(job.ClientIP + "\t" + job.LogStr)
			case <-w.quitChan:
				fmt.Printf("worker%d stopping\n", w.id)
				return
			}
		}
	}()
}

func (w Worker) stop() {
	go func() {
		w.quitChan <- true
	}()
}

// NewDispatcher creates, and returns a new Dispatcher object.
func NewDispatcher(jobQueue chan Job, maxWorkers int) *Dispatcher {
	workerPool := make(chan chan Job, maxWorkers)

	return &Dispatcher{
		jobQueue:   jobQueue,
		maxWorkers: maxWorkers,
		workerPool: workerPool,
	}
}

type Dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
	jobQueue   chan Job
}

func (d *Dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(i+1, d.workerPool)
		worker.start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			go func() {
				fmt.Printf("fetching workerJobQueue\n")
				workerJobQueue := <-d.workerPool
				fmt.Printf("adding\n",)
				workerJobQueue <- job
			}()
		}
	}
}

func requestHandler(w http.ResponseWriter, r *http.Request, jobQueue chan Job, logLocation string) {
	// Make sure we can only be called with an HTTP POST request.
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Set name and validate value.
	logstr := r.FormValue("log")
	if logstr == "" {
		http.Error(w, "You must specify a log.", http.StatusBadRequest)
		return
	}
	var clientIP string
	if len(r.Header.Get("X-Forwarded-For"))==0{
		clientIP = r.RemoteAddr
	} else{
		clientIP = r.Header["X-Forwarded-For"][0]
	}

	// Create Job and push the work onto the jobQueue.
	job := Job{LogStr: logstr, ClientIP: clientIP}
	jobQueue <- job

	// Render success.
	w.WriteHeader(http.StatusCreated)
}

func main() {
	var (
		maxWorkers   = flag.Int("max_workers", 5, "The number of workers to start")
		maxQueueSize = flag.Int("max_queue_size", 100, "The size of job queue")
		port         = flag.String("port", "9999", "The server port")
		configFile = flag.String("c","","Location of the config file")
	)

	flag.Parse()
	error := 0
	if *configFile==""{
        fmt.Println("Required configuration file location. Provide the same using -c <configuration file location>")
        error++
    }
    fmt.Println(error)
	if error > 0{
		return
	}

    logConfig, err := validateConfigFile(*configFile)
    if err!=nil{
    	fmt.Println(err)
    	return
    }
    

    utils.InitLogger(logConfig)

	// Create the job queue.
	jobQueue := make(chan Job, *maxQueueSize)

	// Start the dispatcher.
	dispatcher := NewDispatcher(jobQueue, *maxWorkers)
	dispatcher.run()

	// Start the HTTP handler.
	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		requestHandler(w, r, jobQueue, "/Users/deep.patel/Documents/test_out.log")
	})
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}

func validateConfigFile(configFileLocation string) (*utils.LogConfig, error) {
	config := make(map[string]string)
    err := utils.Load(configFileLocation, config)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%v\n", config)
   	logLocationDir := config["loglocationdir"]
    maxLogFiles,err := strconv.Atoi(config["maxlogfiles"])

	maxFileToDelete,err2 := strconv.Atoi(config["maxfiletodelete"])

	logTrace,err4 := strconv.ParseBool(config["logtrace"])

	var maxSizeOfLogFiles uint32
	
	tempvar, err3 := strconv.Atoi(config["maxsizeoflogfiles"]); 

	if err3 == nil {
	    maxSizeOfLogFiles = uint32(tempvar)
	} else if err3 != nil{
		return nil, err3
	} else{
		if(maxSizeOfLogFiles < 0){
			fmt.Printf("%v\n", config)
			return nil, errors.New("maxSizeOfLogFiles must be greater than 0. maxSizeOfLogFiles=%d")
		}
	}

	if err != nil{
		return nil, err
	} else{
		if maxLogFiles <= 0 || maxLogFiles > 100000 {
			return nil, errors.New("maxfiles must be greater than 0 and less than or equal to 100000: %d")
		}
	}

	if err2 != nil{
		return nil, err2
	} else{
		if maxFileToDelete <= 0 || maxFileToDelete > maxLogFiles {
			return nil, errors.New("nfilesToDel must be greater than 0 and less than or equal to maxfiles! toDel=%d maxfiles=%d")
		}
	}
	
	if err4 != nil{
		return nil, err4
	}

	logConfig := &utils.LogConfig{LogLocationDir: logLocationDir,
						MaxLogFiles: maxLogFiles,
						MaxFileToDelete: maxFileToDelete,
						MaxSizeOfLogFiles: maxSizeOfLogFiles,
						LogTrace: logTrace}

	return logConfig, nil
}