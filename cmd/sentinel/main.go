package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/controller"
	"github.com/ayuspoudel/sentinel-sre/internal/metrics"
	"github.com/ayuspoudel/sentinel-sre/internal/prometheus"
	"github.com/ayuspoudel/sentinel-sre/internal/server"
)

func main() {
	srv := server.New(":8000")
	metrics.Register()
	/*
		Initializing prom client using our New() function in prometheus/client
	*/
	prom := prometheus.New("http://localhost:9090")
	/*
		initializing decision engine (controller) which is responsible for evalating
		system health and deciding whether blocking deployments should be allowed
		or blocked. The error rate 0.01 represents, policy not logic, can differ across
		envs.
	*/
	ctrl := controller.New(prom, 0.01)

	/*
		Context controlling the controller lifecycle.
		This context is used to ensure the controller stops evaluating
		when the application is shutting down, preventing background
		goroutines from leaking or running after shutdown begins.
	*/
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ctrl.Start(ctx)
	srv.HandleFunc("/status", ctrl.StatusHandler)
	/*
		@ayuspoudel
		Go routines are extermemly helpful in cases where we call functions like
		http.ListenAndServe(), which blocks for entire lifetime of server. Running
		the server in a go routine would allow us to run http server concurrently
		and main go routine to continue executing. The program will this listen for
		all OS signals for termination.
	*/
	go func() {
		err := srv.Start()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %+v", err)
		}
	}()

	/*
		@ayuspoudel
		The main go routine will wait for any signal from the OS to stop the server, i.e
		handle the process lifecycle. If it recieves any SIGINT or SIGTERM from OS, it will
		signal.Notify instruct go to forward those signals into 'stop' channel.
		PS. reciever operation (<-stop) blocks the main go routine until a termination has occured
	*/
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	/*
		@ayuspoudel
		Once a termination signal is received, we initiate graceful shutdown.A context with
		timeout is created to enforce an upper bound on how long the server is allowed to shut down.
		The timeot (10 seconds here) represents the maximum time
		the server may take to:
		- stop accepting new connections
		- finish in-flight HTTP requests
		- release resources cleanly
		If the timeout expires, shutdown is forced.
	*/
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	log.Println("server exited cleanly")

}

/*
	@ayuspoudel

	A little note for myself
	========================================================================
	HOW KUBERNETES INTERACTS WITH THIS CODE AND WHY THIS PATTERN IS CRITICAL
	========================================================================

	This program follows a very specific server lifecycle pattern that is
	required for running safely inside Kubernetes (and similar container
	orchestrators). Kubernetes does NOT understand goroutines, HTTP servers,
	or Go internals. It only understands processes, signals, and time.
	When this Go program runs inside a container, it becomes PID 1. Kubernetes
	manages the lifecycle of this process entirely through OS signals.

	SERVER STARTUP PHASE
	When the container starts, main() is executed. The HTTP server is started
	inside a goroutine because srv.Start() internally calls
	http.ListenAndServe(), which blocks for the entire lifetime of the server.
	If srv.Start() were called directly (without a goroutine), the main
	goroutine would block forever and never reach the shutdown logic.
	This would make graceful shutdown impossible.
	The goroutine exists ONLY to run the blocking server loop while allowing
	the main goroutine to continue executing. It does not signal anything,
	it does not handle OS signals, and it does not control shutdown.


	RUNNING STATE
	While the server is running:
	- The goroutine is blocked inside ListenAndServe(), handling HTTP traffic
	- The main goroutine is blocked on `<-stop`, waiting for an OS signal
	Both run concurrently. There is no communication between them.


	KUBERNETES SHUTDOWN SEQUENCE
	When Kubernetes wants to stop the pod (during deployments, scaling,
	node drain, or pod deletion), it follows a strict sequence:

	1. Kubernetes sends SIGTERM to the process (PID 1)
	2. Kubernetes waits for terminationGracePeriodSeconds (default 30s)
	3. If the process is still running, Kubernetes sends SIGKILL

	SIGKILL cannot be handled or intercepted. The process is terminated
	immediately and forcefully.
	SIGNAL HANDLING IN THIS CODE

	The line:
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	instructs Go to forward SIGINT and SIGTERM signals from the OS into
	the `stop` channel.

	The line:
		<-stop
	blocks the main goroutine until Kubernetes (or the OS) sends SIGTERM.
	This is the ONLY trigger for graceful shutdown.

	The goroutine running the server is NOT involved in signal handling.

	GRACEFUL SHUTDOWN PHASE
	Once SIGTERM is received, the program creates a context with a timeout
	of 10 seconds. This timeout defines the maximum amount of time the server
	is allowed to shut down cleanly.

	The call:
		srv.Shutdown(ctx)
	does the following internally:
	- Stops accepting new connections
	- Allows existing connections to continue
	- Waits for in-flight requests to finish
	- Exits early if all requests complete
	- Forces shutdown if the timeout expires

	The timeout is critical. Without it, shutdown could hang forever and
	Kubernetes would eventually send SIGKILL.


	WHAT HAPPENS IF THIS PATTERN IS NOT USED
	Case 1: Server started without a goroutine
	- main() blocks forever in ListenAndServe()
	- SIGTERM is never handled
	- Kubernetes waits, then sends SIGKILL
	- Active requests are dropped mid-flight
	- Clients see connection resets

	Case 2: No signal handling
	- SIGTERM is ignored
	- Kubernetes eventually sends SIGKILL
	- Logs may not flush
	- Metrics and traces are lost
	- Requests fail unpredictably

	Case 3: No context timeout
	- Shutdown can hang indefinitely
	- Kubernetes grace period expires
	- SIGKILL terminates the process
	- Same crash behavior as above


	WHY THIS CAUSES REAL PRODUCTION CRASHES
	Without graceful shutdown:
	- Load balancers still route traffic
	- Pods disappear mid-request
	- Clients retry aggressively
	- Retry storms occur
	- Latency spikes
	- Error rates increase
	- Cascading failures can happen

	This is how seemingly small mistakes turn into production incidents.


	FINAL SUMMARY
	Kubernetes communicates ONLY through OS signals and timeouts.
	This pattern exists to ensure:
	- SIGTERM is handled
	- Traffic drains cleanly
	- In-flight requests complete
	- The process exits before SIGKILL

	The goroutine enables concurrency.
	The signal handling enables lifecycle control.
	The context timeout enforces safety.

	Without all three, crashes are inevitable.
*/
