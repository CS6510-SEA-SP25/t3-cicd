/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	hpav1 "cicd.operator/hpa/api/v1"
)

type QueueItemWithDependency struct {
	Id         string              `json:"id"`
	Dependency map[string][]string `json:"dependency"`
}

// RabbitMQConnection holds the connection and channel
type RabbitMQConnection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// RedisConnection holds the Redis client connection
type RedisConnection struct {
	Client  *redis.Client
	Context context.Context
}

// PoolScalerReconciler reconciles a PoolScaler object
type PoolScalerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Clientset *kubernetes.Clientset
}

// +kubebuilder:rbac:groups=hpa.cicd.operator,resources=poolscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hpa.cicd.operator,resources=poolscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hpa.cicd.operator,resources=poolscalers/finalizers,verbs=update
// +kubebuilder:rbac:groups=hpa.cicd.operator,resources=services;endpoints,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PoolScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *PoolScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Fetch the PoolScaler instance
	// The purpose is check if the Custom Resource for the Kind PoolScaler
	// is applied on the cluster if not we return nil to stop the reconciliation
	poolScaler := &hpav1.PoolScaler{}
	if err := r.Get(ctx, req.NamespacedName, poolScaler); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get RabbitMQ message count
	password, err := r.getSecretValue(poolScaler.Namespace, poolScaler.Spec.InputQueue.PasswordSecretRef)
	if err != nil {
		return r.handleError(ctx, poolScaler, "FailedToGetPassword", err)
	}

	messageCount, err := r.getQueueMessageCount(poolScaler, password)
	if err != nil {
		return r.handleError(ctx, poolScaler, "QueueCheckFailed", err)
	}

	// Update status with current queue size
	poolScaler.Status.QueueMessages = messageCount
	if err := r.Status().Update(ctx, poolScaler); err != nil {
		return ctrl.Result{}, err
	}

	return r.reconcileDeploymentScaling(ctx, poolScaler, messageCount, password)
}

func (r *PoolScalerReconciler) reconcileDeploymentScaling(ctx context.Context, poolScaler *hpav1.PoolScaler, messageCount int32, password string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 1. Establish RabbitMQ connection
	rmqConn, err := r.connectRabbitMQ(poolScaler, password)
	if err != nil {
		return r.handleError(ctx, poolScaler, "RabbitMQConnectionFailed", err)
	}
	defer r.closeRabbitMQ(rmqConn)

	// 2. Establish Redis connection
	// redisPassword, err := r.getRedisPassword(ctx, poolScaler)
	// if err != nil {
	// 	return r.handleError(ctx, poolScaler, "GetRedisPasswordFailed", err)
	// }

	// redisConn, err := r.connectRedis(poolScaler, redisPassword)
	// if err != nil {
	// 	return r.handleError(ctx, poolScaler, "RedisConnectionFailed", err)
	// }
	// defer r.closeRedis(redisConn)

	// 3. Process new messages polling from the queue
	for i := int32(0); i < messageCount; i++ {
		// Get message from queue
		msg, err := r.consumeMessage(rmqConn, poolScaler.Spec.InputQueue.QueueName)
		if err != nil {
			log.Error(err, "Failed to get message")
			continue
		}
		if msg == nil {
			break // No more messages
		}

		// Process one message
		if err := r.processSingleMessage(ctx, poolScaler, rmqConn.Channel, msg, i); err != nil {
			log.Error(err, "Failed to process message")
			continue
		}
	}

	return ctrl.Result{
		RequeueAfter: time.Duration(poolScaler.Spec.Scaling.PollingIntervalSeconds) * time.Second,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PoolScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hpav1.PoolScaler{}).
		Named("poolscaler").
		Complete(r)
}

func (r *PoolScalerReconciler) handleError(ctx context.Context, poolScaler *hpav1.PoolScaler, reason string, err error) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	poolScaler.Status.Phase = "Error"
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            err.Error(),
		LastTransitionTime: metav1.Now(),
	}
	poolScaler.Status.Conditions = append(poolScaler.Status.Conditions, condition)

	if updateErr := r.Status().Update(ctx, poolScaler); updateErr != nil {
		log.Error(updateErr, "Failed to update status")
		return ctrl.Result{}, updateErr
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, err
}

// Fetch svc Kubernetes endpoint of RabbitMQ
func (r *PoolScalerReconciler) getServiceEndpoint(ctx context.Context, serviceHost string, instance *hpav1.PoolScaler) (string, error) {
	// log.Printf("getServiceEndpoint %s", serviceHost)
	// If host is already a full URL, use it directly
	if strings.Contains(serviceHost, ".") {
		return serviceHost, nil
	}

	// Get service details
	svc, err := r.Clientset.CoreV1().Services(instance.Namespace).Get(
		ctx,
		serviceHost, // "pipeline-queue" or "job-queue" or "minio"
		metav1.GetOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get RabbitMQ service: %w", err)
	}

	// Handle different service types
	switch svc.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		return fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace), nil
	case corev1.ServiceTypeLoadBalancer:
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return "", fmt.Errorf("LoadBalancer IP not yet allocated")
		}
		return svc.Status.LoadBalancer.Ingress[0].IP, nil
	case corev1.ServiceTypeNodePort:
		return fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace), nil
	default:
		return "", fmt.Errorf("unsupported service type: %s", svc.Spec.Type)
	}
}

// Get secret values
func (r *PoolScalerReconciler) getSecretValue(namespace string, secretRef hpav1.SecretReference) (string, error) {
	secret := &corev1.Secret{}
	err := r.Get(context.Background(),
		types.NamespacedName{
			Name:      secretRef.Name,
			Namespace: namespace,
		},
		secret,
	)
	if err != nil {
		return "", err
	}

	value, ok := secret.Data[secretRef.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", secretRef.Key, secretRef.Name)
	}

	return string(value), nil
}

// Gets the current message count in the queue
func (r *PoolScalerReconciler) getQueueMessageCount(poolScaler *hpav1.PoolScaler, password string) (int32, error) {
	// Establish connection
	rmq, err := r.connectRabbitMQ(poolScaler, password)
	if err != nil {
		return 0, err
	}
	defer r.closeRabbitMQ(rmq)

	// QueueDeclare declares a queue to hold messages and deliver to consumers.
	// Declaring creates a queue if it doesn't already exist,
	// or ensures that an existing queue matches the same parameters.
	queue, err := rmq.Channel.QueueDeclare(
		poolScaler.Spec.InputQueue.QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return 0, fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Printf("getQueueMessageCount %v", queue.Messages)
	return int32(queue.Messages), nil
}

// Establishes a new connection to RabbitMQ
func (r *PoolScalerReconciler) connectRabbitMQ(poolScaler *hpav1.PoolScaler, password string) (*RabbitMQConnection, error) {
	host, err := r.getServiceEndpoint(context.Background(), poolScaler.Spec.InputQueue.Host, poolScaler)
	if err != nil {
		return nil, err
	}

	connString := fmt.Sprintf("amqp://%s:%s@%s:%d",
		poolScaler.Spec.InputQueue.Username,
		password,
		host,
		poolScaler.Spec.InputQueue.Port,
	)

	conn, err := amqp.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close() // Close connection if channel creation fails
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQConnection{
		Conn:    conn,
		Channel: ch,
	}, nil
}

// Safely closes the connection and channel
func (r *PoolScalerReconciler) closeRabbitMQ(rmq *RabbitMQConnection) {
	if rmq == nil {
		return
	}

	if rmq.Channel != nil {
		rmq.Channel.Close()
	}
	if rmq.Conn != nil {
		rmq.Conn.Close()
	}
}

// Consumes a single message from the queue
func (r *PoolScalerReconciler) consumeMessage(rmq *RabbitMQConnection, queueName string) (*amqp.Delivery, error) {
	msg, ok, err := rmq.Channel.Get(
		queueName,
		false, // auto-ack
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	if !ok {
		return nil, nil // No messages available
	}
	return &msg, nil
}

// Handles creation of pod for a single message
func (r *PoolScalerReconciler) processSingleMessage(ctx context.Context, poolScaler *hpav1.PoolScaler, ch *amqp.Channel, msg *amqp.Delivery, index int32) error {
	log := logf.FromContext(ctx)

	// Create pod
	pod := r.createWorkerPod(poolScaler, index, msg.Body)
	if err := r.Create(ctx, pod); err != nil {
		// Requeue message on failure
		if nackErr := ch.Nack(msg.DeliveryTag, false, true); nackErr != nil {
			log.Error(nackErr, "Failed to requeue message after pod creation failure")
		}
		return fmt.Errorf("failed to create pod: %w", err)
	}

	// Acknowledge message
	if err := ch.Ack(msg.DeliveryTag, false); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	log.Info("Successfully processed message", "pod", pod.Name, "messageLength", len(msg.Body))
	return nil
}

// Create a worker pod to execute one pipeline
func (r *PoolScalerReconciler) createWorkerPod(poolScaler *hpav1.PoolScaler, index int32, messageBody []byte) *corev1.Pod {
	// Get service endpoints
	inputQueueHost, err := r.getServiceEndpoint(context.Background(), poolScaler.Spec.InputQueue.Host, poolScaler)
	if err != nil {
		logf.FromContext(context.Background()).Error(err, "Failed to get InputQueue Endpoint")
	}
	storageHost, err := r.getServiceEndpoint(context.Background(), poolScaler.Spec.Storage.Host, poolScaler)
	if err != nil {
		logf.FromContext(context.Background()).Error(err, "Failed to get MinIO Endpoint")
	}

	podName := fmt.Sprintf("%s-worker-%d-%d", poolScaler.Name, time.Now().Unix(), index)

	// Get secrets
	inputQueuePassword, _ := r.getSecretValue(poolScaler.Namespace, poolScaler.Spec.InputQueue.PasswordSecretRef)
	dbPassword, _ := r.getSecretValue(poolScaler.Namespace, poolScaler.Spec.Database.PasswordSecretRef)
	storageAccessKey, _ := r.getSecretValue(poolScaler.Namespace, poolScaler.Spec.Storage.AccessKeyRef)
	storageSecretKey, _ := r.getSecretValue(poolScaler.Namespace, poolScaler.Spec.Storage.SecretKeyRef)
	cachePassword, _ := r.getSecretValue(poolScaler.Namespace, poolScaler.Spec.Cache.PasswordSecretRef)

	// Worker environment variables
	envVars := []corev1.EnvVar{
		// InputQueue
		{
			Name: "RABBITMQ_URL",
			Value: fmt.Sprintf("amqp://%s:%s@%s:%d",
				poolScaler.Spec.InputQueue.Username,
				inputQueuePassword,
				inputQueueHost,
				poolScaler.Spec.InputQueue.Port,
			),
		},
		{Name: "TASK_QUEUE", Value: poolScaler.Spec.InputQueue.QueueName},
		// Database
		{Name: "DB_HOST", Value: poolScaler.Spec.Database.Host},
		{Name: "DB_PORT", Value: fmt.Sprintf("%d", poolScaler.Spec.Database.Port)},
		{Name: "DB_USER", Value: poolScaler.Spec.Database.Username},
		{Name: "DB_NAME", Value: poolScaler.Spec.Database.Name},
		{Name: "DB_SSL_MODE", Value: poolScaler.Spec.Database.SSLMode},
		{Name: "DB_PASSWORD", Value: dbPassword},
		{Name: "DB_SSL_CA", Value: "/etc/ssl/certs/ca.pem"},
		// Storage
		{Name: "MINIO_ENDPOINT", Value: storageHost},
		{Name: "MINIO_ACCESS_KEY", Value: storageAccessKey},
		{Name: "MINIO_SECRET_KEY", Value: storageSecretKey},
		{Name: "DEFAULT_BUCKET", Value: poolScaler.Spec.Storage.DefaultBucket},
		// Cache
		{Name: "REDIS_HOST", Value: poolScaler.Spec.Cache.Host},
		{Name: "REDIS_PORT", Value: fmt.Sprintf("%d", poolScaler.Spec.Cache.Port)},
		{Name: "REDIS_USERNAME", Value: poolScaler.Spec.Cache.Username},
		{Name: "REDIS_PASSWORD", Value: cachePassword},
	}

	// OutputQueue
	if !reflect.DeepEqual(poolScaler.Spec.OutputQueue, hpav1.RabbitMQConfig{}) {
		outputQueuePassword, _ := r.getSecretValue(poolScaler.Namespace, poolScaler.Spec.OutputQueue.PasswordSecretRef)

		envVars = append(envVars, corev1.EnvVar{
			Name: "JOB_QUEUE_URL",
			Value: fmt.Sprintf("amqp://%s:%s@%s:%d",
				poolScaler.Spec.OutputQueue.Username,
				outputQueuePassword,
				poolScaler.Spec.OutputQueue.Host,
				poolScaler.Spec.OutputQueue.Port,
			),
		})
		envVars = append(envVars, corev1.EnvVar{Name: "JOB_QUEUE_NAME", Value: poolScaler.Spec.OutputQueue.QueueName})
	}

	// Docker Volume
	var hostPathType corev1.HostPathType = corev1.HostPathSocket
	dockerVolume := corev1.Volume{
		Name: "docker-socket",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/run/docker.sock", // Path on host to Docker socket
				Type: &hostPathType,
			},
		},
	}
	dockerVolumeMount := corev1.VolumeMount{
		Name:      "docker-socket",
		MountPath: "/var/run/docker.sock", // Mount Docker socket inside container
	}

	// SSL Volume
	sslVolume := corev1.Volume{
		Name: "ca-cert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: poolScaler.Spec.Database.SSLCASecretRef.Name,
				Items: []corev1.KeyToPath{
					{
						Key:  poolScaler.Spec.Database.SSLCASecretRef.Key,
						Path: "ca.pem",
					},
				},
			},
		},
	}
	sslVolumeMount := corev1.VolumeMount{
		Name:      "ca-cert",
		MountPath: "/etc/ssl/certs/ca.pem",
		SubPath:   "ca.pem",
	}

	// Arguments
	args := []string{"--input", string(messageBody)}

	// Create pod with SSL and Docker volumes
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: poolScaler.Namespace,
			Labels: map[string]string{
				"app":                 "poolscaler-worker",
				"poolscaler-instance": poolScaler.Name,
				"pod-restart-policy":  "never", // never restart
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:         "worker",
					Image:        fmt.Sprintf("%s:%s", poolScaler.Spec.WorkerPool.WorkerImage, poolScaler.Spec.WorkerPool.WorkerImageTag),
					Args:         args,
					Env:          envVars,
					VolumeMounts: []corev1.VolumeMount{dockerVolumeMount, sslVolumeMount},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),  // 0.5 CPU
							corev1.ResourceMemory: resource.MustParse("256Mi"), // 256MB memory
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),  // 0.5 CPU
							corev1.ResourceMemory: resource.MustParse("512Mi"), // 256GB memory
						},
					},
				},
			},
			Volumes:                       []corev1.Volume{sslVolume, dockerVolume},
			TerminationGracePeriodSeconds: ptr.To(int64(30)),
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(poolScaler, pod, r.Scheme); err != nil {
		logf.FromContext(context.Background()).Error(err, "Failed to set owner reference")
	}

	return pod
}
