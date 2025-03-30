//nolint:all
package k8s

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// k8s config
var kubeconfig *string

// Create a local random generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func init() {
	// Ensure the flag is defined only once
	if flag.Lookup("kubeconfig") == nil {
		kubeconfig = flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "Path to kubeconfig file")
	}

	// Only parse flags if not running in a test context
	if !isTestMode() {
		flag.Parse()
	}
}

// isTestMode checks if the code is running in a test context.
func isTestMode() bool {
	for _, arg := range os.Args {
		if arg == "-test.v" || arg == "-test.run" || arg == "-test.paniconexit0" {
			return true
		}
	}
	return false
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rng.Intn(len(letterRunes))]
	}
	return string(b)
}

// LoadKubeConfig loads the kubeconfig file to configure the connection to the Kubernetes cluster.
func loadKubeConfig() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// creates the in-cluster config
	// config, err = rest.InClusterConfig()

	// Check if running inside Kubernetes
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster {
		fmt.Println("Running inside Kubernetes cluster...")
		config, err = rest.InClusterConfig()
	} else {
		fmt.Println("Running outside Kubernetes, using kubeconfig...")

		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}

	if err != nil {
		panic(err.Error())
	}

	// Create the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}

// createPod creates a new pod in the specified namespace with the given parameters.
func createPod(clientset *kubernetes.Clientset, podName, namespace, image string, args []string) (*corev1.Pod, error) {
	// Set pod's env vars
	var envVars = map[string]string{
		// General
		"DB_HOST":      os.Getenv("DB_HOST"),
		"DB_PORT":      os.Getenv("DB_PORT"),
		"DB_USER":      os.Getenv("DB_USER"),
		"DB_PASSWORD":  os.Getenv("DB_PASSWORD"),
		"DB_NAME":      os.Getenv("DB_NAME"),
		"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
		// MinIO
		"MINIO_ENDPOINT":   os.Getenv("MINIO_ENDPOINT"),
		"MINIO_ACCESS_KEY": os.Getenv("MINIO_ACCESS_KEY"),
		"MINIO_SECRET_KEY": os.Getenv("MINIO_SECRET_KEY"),
		"DEFAULT_BUCKET":   os.Getenv("DEFAULT_BUCKET"),
		// RabbitMQ
		"RABBITMQ_URL": os.Getenv("RABBITMQ_URL"),
		"TASK_QUEUE":   os.Getenv("TASK_QUEUE"),
		// Redis
		"REDIS_HOST":     os.Getenv("REDIS_HOST"),
		"REDIS_PORT":     os.Getenv("REDIS_PORT"),
		"REDIS_USERNAME": os.Getenv("REDIS_USERNAME"),
		"REDIS_PASSWORD": os.Getenv("REDIS_PASSWORD"),
	}
	var envs []corev1.EnvVar
	for key, value := range envVars {
		envs = append(envs, corev1.EnvVar{Name: key, Value: value})
	}

	// Define volume for mounting Docker socket
	volumeMount := corev1.VolumeMount{
		Name:      "docker-socket",
		MountPath: "/var/run/docker.sock", // Mount Docker socket inside container
	}

	// Define volume to be mounted
	var hostPathType corev1.HostPathType = corev1.HostPathSocket
	volume := corev1.Volume{
		Name: "docker-socket",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/run/docker.sock", // Path on host to Docker socket
				Type: &hostPathType,
			},
		},
	}

	// Define pod specification
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:         "worker-api-container",
					Image:        image,
					Args:         args,
					Env:          envs,
					VolumeMounts: []corev1.VolumeMount{volumeMount},
					// Resources: corev1.ResourceRequirements{

					// },
				},
			},
			Volumes: []corev1.Volume{volume}, // Attach volume to the pod
		},
	}

	// Create the pod using the clientset
	podClient := clientset.CoreV1().Pods(namespace)
	createdPod, err := podClient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	return createdPod, nil
}

// PrintPodDetails prints the details of the created pod.
func printPodDetails(pod *corev1.Pod) {
	fmt.Printf("Pod created successfully:\n")
	fmt.Printf("Name: %s\n", pod.Name)
	fmt.Printf("Namespace: %s\n", pod.Namespace)
	fmt.Printf("Containers: %v\n", pod.Spec.Containers)
}

/* Create a worker instance and execute with given arguments */
func CreateWorkerInstance(args []string) {
	// Define your arguments and image for the container
	podName := "worker-api-pod-" + RandStringRunes(8)
	namespace := "default"
	image := "minh160302/worker-api" // Specify your image, like in `docker run worker-api`

	// Load the kubeconfig and create the clientset
	clientset, err := loadKubeConfig()
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	// Create the pod with arguments, just like `docker run worker-api <arguments>`
	createdPod, err := createPod(clientset, podName, namespace, image, args)
	if err != nil {
		log.Fatalf("Error creating pod: %v", err)
	}

	// Print the pod details
	printPodDetails(createdPod)
}
