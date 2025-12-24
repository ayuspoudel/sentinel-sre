package main

import (
	"bufio"
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var allClusters = []string{"paymentsCluster", "streamingCluster", "sreCluster"}
var singleCluster = []string{"sreCluster"}

var kubeMu sync.Mutex

func checkBinary(name string) {
	if _, err := exec.LookPath(name); err != nil {
		log.Fatalf("%s is not installed", name)
	}
}

func run(ctx context.Context, cmd string, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func output(cmd string, args ...string) string {
	c := exec.Command(cmd, args...)
	var out bytes.Buffer
	c.Stdout = &out
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func kubectl(ctx context.Context, kubeCtx string, args ...string) error {
	kubeMu.Lock()
	defer kubeMu.Unlock()
	return run(ctx, "kubectl", append([]string{"--context", kubeCtx}, args...)...)
}

func kubectlOut(ctx context.Context, kubeCtx string, args ...string) string {
	kubeMu.Lock()
	defer kubeMu.Unlock()
	return output("kubectl", append([]string{"--context", kubeCtx}, args...)...)
}

func waitFor(ctx context.Context, interval, timeout time.Duration, fn func() bool) bool {
	t := time.NewTicker(interval)
	defer t.Stop()

	timer := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return false
		case <-timer:
			return false
		case <-t.C:
			if fn() {
				return true
			}
		}
	}
}

func loadEnv() {
	var envPath string
	_, srcFile, _, ok := runtime.Caller(0)
	if ok {
		srcDir := filepath.Dir(srcFile)
		candidate := filepath.Join(srcDir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			envPath = candidate
		}
	}

	if envPath == "" {
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			candidate := filepath.Join(exeDir, ".env")
			if _, err := os.Stat(candidate); err == nil {
				envPath = candidate
			}
		}
	}

	if envPath == "" {
		log.Fatal(".env file not found (checked source dir and executable dir)")
	}

	f, err := os.Open(envPath)
	if err != nil {
		log.Fatalf("failed to open .env at %s", envPath)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) == 2 {
			os.Setenv(kv[0], kv[1])
		}
	}

	log.Printf("loaded env from %s\n", envPath)
}

func startClustersSequential(ctx context.Context, clusters []string) error {
	for _, c := range clusters {
		if err := run(ctx, "minikube", "start", "-p", c, "--driver=docker"); err != nil {
			return err
		}
	}
	return nil
}

func installPrometheusCRDs(ctx context.Context, kubeCtx string) error {
	crdURL := "https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.76.2/stripped-down-crds.yaml"
	return kubectl(ctx, kubeCtx, "apply", "-f", crdURL)
}

func portForwardPrometheus(ctx context.Context, kubeCtx string, localPort string) {
	for {
		if ctx.Err() != nil {
			return
		}

		log.Printf("[%s] prometheus port-forward\n", kubeCtx)

		cmd := exec.CommandContext(
			ctx,
			"kubectl", "--context", kubeCtx,
			"-n", "monitoring",
			"port-forward",
			"svc/prometheus-k8s",
			localPort+":9090",
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()

		time.Sleep(2 * time.Second)
	}
}

func argocdHasInsecure(ctx context.Context, kubeCtx string) bool {
	out := kubectlOut(ctx, kubeCtx,
		"-n", "argocd",
		"get", "deployment", "argocd-server",
		"-o", "jsonpath={.spec.template.spec.containers[0].args}",
	)
	return strings.Contains(out, "--insecure")
}

func patchArgoCD(ctx context.Context, kubeCtx string) {
	kubectl(ctx, kubeCtx,
		"patch", "configmap", "argocd-cm",
		"-n", "argocd",
		"--type", "merge",
		"-p", `{"data":{"server.insecure":"true"}}`,
	)

	if !argocdHasInsecure(ctx, kubeCtx) {
		kubectl(ctx, kubeCtx,
			"patch", "deployment", "argocd-server",
			"-n", "argocd",
			"--type", "json",
			"-p", `[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--insecure"}]`,
		)
	}

	kubectl(ctx, kubeCtx,
		"rollout", "restart", "deployment/argocd-server",
		"-n", "argocd",
	)
}

func installArgoCD(ctx context.Context, kubeCtx string) {
	run(ctx, "bash", "-c",
		"kubectl --context "+kubeCtx+" create ns argocd --dry-run=client -o yaml | kubectl --context "+kubeCtx+" apply -f -")

	kubectl(ctx, kubeCtx, "apply", "-n", "argocd",
		"-f", "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml")

	waitFor(ctx, 2*time.Second, 3*time.Minute, func() bool {
		return kubectl(ctx, kubeCtx, "-n", "argocd", "get", "deployment", "argocd-server") == nil
	})

	patchArgoCD(ctx, kubeCtx)
}

func portForwardArgoCD(ctx context.Context, kubeCtx string) {
	for {
		if ctx.Err() != nil {
			return
		}

		log.Println("[argocd] port-forward")

		cmd := exec.CommandContext(
			ctx,
			"kubectl", "--context", kubeCtx,
			"-n", "argocd",
			"port-forward",
			"svc/argocd-server",
			"8088:80",
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()

		time.Sleep(2 * time.Second)
	}
}

func startNgrok(ctx context.Context, port string, name string) {
	token := os.Getenv("NGROK_AUTHTOKEN")
	if token == "" {
		log.Fatal("NGROK_AUTHTOKEN not set")
	}

	run(ctx, "ngrok", "config", "add-authtoken", token)

	cmd := exec.CommandContext(ctx, "ngrok", "http", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func waitForArgoCDApps(ctx context.Context, kubeCtx string) {
	log.Println("[argocd] waiting for applications to become healthy")

	ok := waitFor(ctx, 5*time.Second, 10*time.Minute, func() bool {
		out := kubectlOut(ctx, kubeCtx,
			"-n", "argocd",
			"get", "applications",
			"-o", "jsonpath={range .items[*]}{.status.health.status}{\" \"}{end}",
		)
		if out == "" {
			return false
		}

		for _, s := range strings.Fields(out) {
			if s != "Healthy" {
				return false
			}
		}
		return true
	})

	if !ok {
		log.Fatal("argocd applications did not become healthy")
	}

	log.Println("[argocd] all applications healthy")
}

func main() {
	log.SetFlags(0)

	checkBinary("minikube")
	checkBinary("kubectl")
	checkBinary("ngrok")

	loadEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	activeClusters := allClusters
	if len(os.Args) > 1 && os.Args[1] == "--single" {
		activeClusters = singleCluster
	}

	if err := startClustersSequential(ctx, activeClusters); err != nil {
		log.Fatal(err)
	}

	for _, c := range activeClusters {
		if err := installPrometheusCRDs(ctx, c); err != nil {
			log.Fatalf("prometheus CRDs failed on %s", c)
		}
	}

	installArgoCD(ctx, "sreCluster")

	// Expose Argo CD only
	go portForwardArgoCD(ctx, "sreCluster")
	go startNgrok(ctx, "8088", "argocd")

	// Block until Helm charts (including Prometheus) are deployed
	waitForArgoCDApps(ctx, "sreCluster")

	// Prometheus exists now → safe to port-forward locally
	port := 9091
	for _, c := range activeClusters {
		p := port
		go portForwardPrometheus(ctx, c, strconv.Itoa(p))
		port++
	}

	<-ctx.Done()
}
