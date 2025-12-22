package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
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

func waitFor(ctx context.Context, fn func() bool) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if fn() {
				return
			}
		}
	}
}

func startClustersSequential(ctx context.Context, clusters []string) error {
	for _, c := range clusters {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := run(ctx, "minikube", "start", "-p", c, "--driver=docker"); err != nil {
			return err
		}
	}
	return nil
}

func installArgoCD(ctx context.Context, kubeCtx string) {
	if ctx.Err() != nil {
		return
	}

	run(ctx, "bash", "-c",
		"kubectl --context "+kubeCtx+" create namespace argocd --dry-run=client -o yaml | kubectl --context "+kubeCtx+" apply -f -")

	kubectl(ctx, kubeCtx, "apply", "-n", "argocd",
		"-f", "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml")

	waitFor(ctx, func() bool {
		return kubectl(ctx, kubeCtx, "-n", "argocd", "get", "deployment", "argocd-server") == nil
	})

	waitFor(ctx, func() bool {
		return kubectl(ctx, kubeCtx, "-n", "argocd", "wait",
			"--for=condition=available",
			"deployment/argocd-server",
			"--timeout=1s") == nil
	})

	waitFor(ctx, func() bool {
		return kubectl(ctx, kubeCtx, "-n", "argocd", "get", "secret", "argocd-initial-admin-secret") == nil
	})

	password := kubectlOut(ctx, kubeCtx,
		"-n", "argocd",
		"get", "secret", "argocd-initial-admin-secret",
		"-o", "jsonpath={.data.password}",
	)

	if password != "" {
		log.Println(output("bash", "-c", "echo "+password+" | base64 --decode"))
	}

	waitFor(ctx, func() bool {
		return kubectl(ctx, kubeCtx, "kubectl", "apply", "-n", "argocd", "-f", "../argocd/root.yaml") == nil
	})
}

func portForward(ctx context.Context, kubeCtx string) {
	cmd := exec.CommandContext(
		ctx,
		"kubectl", "--context", kubeCtx,
		"-n", "argocd",
		"port-forward", "svc/argocd-server",
		"8088:443",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func cleanup(clusters []string) {
	for _, c := range clusters {
		_ = exec.Command("minikube", "delete", "-p", c).Run()
	}
}

func main() {
	log.SetFlags(0)
	checkBinary("minikube")
	checkBinary("kubectl")

	useSingle := false
	if len(os.Args) > 1 && os.Args[1] == "--single" {
		useSingle = true
	}

	activeClusters := allClusters
	if useSingle {
		activeClusters = singleCluster
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer cleanup(activeClusters)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	if err := startClustersSequential(ctx, activeClusters); err != nil {
		return
	}

	if ctx.Err() != nil {
		return
	}

	installArgoCD(ctx, "sreCluster")

	if ctx.Err() != nil {
		return
	}

	go portForward(ctx, "sreCluster")
	<-ctx.Done()
}
