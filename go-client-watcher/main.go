package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func main() {
	if err := doMain(); err != nil {
		panic(err.Error())
	}
}

func doMain() error {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return fmt.Errorf("error loading Kuberentes config for Minikube: %w", err)
	}

	//// Load Kubernetes config
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	//	if err != nil {
	//		return fmt.Errorf("error loading Kuberentes config: %w", err)
	//	}
	//}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %w", err)
	}

	podClient := clientset.CoreV1().Pods("")
	watcher, err := podClient.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error creating Pod watcher: %w", err)
	}

	// Handle signals for graceful shutdown
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stopCh)

	// Process events from the watcher
	for {
		select {
		case event := <-watcher.ResultChan():

			// Check if pod event
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			var action string
			switch event.Type {
			case watch.Added:
				action = "created"
			case watch.Modified:
				action = "modified"
			case watch.Deleted:
				action = "deleted"
			case watch.Error:
				return fmt.Errorf("expected valid event.Type, but got an error: %w", watch.Error)
			default:
				return fmt.Errorf("expected valid event.Type, but got %v", event.Type)
			}

			fmt.Printf("Pod %s (%s): %s\n", action, pod.Namespace, pod.Name)

		case <-stopCh:
			fmt.Println("Shutting down Pod watcher...")
			return nil
		}
	}
}
