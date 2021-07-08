package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	pb "github.com/accuknox/KubeArmor/protobuf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
totalAlertsRequestsinHost = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_in_host_total",
		Help: "Total number of logs generated from Kubearmor Relay based on HostName",
	}, []string{"HostName"})


totalAlertsRequestsinPod = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_in_pod_total",
		Help: "Total number of logs generated from Kubearmor Relay based on PodName",
	}, []string{"PodName"})


totalAlertsRequestsinNamespace = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_in_namespace_total",
		Help: "Total number of logs generated from Kubearmor Relay based on Namespace",
	}, []string{"NamespaceName"})


totalAlertsRequestsinContainer = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_in_container_total",
		Help: "Total number of logs generated from Kubearmor Relay based on Container",
	}, []string{"ContainerName"})

totalAlertsWithPolicy = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_policy_logs_total",
		Help: "Total number of logs generated on a given policy",
	}, []string{"PolicyName"})

totalAlertsWithSeverity = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_with_severity_total",
		Help: "Total number of logs generated with X severity or above",
	}, []string{"Severity"})

totalAlertsWithType = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_with_type_total",
		Help: "Total number of logs generated from Kubearmor Relay based on given type",
	}, []string{"Type"})

totalAlertsWithOperation = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_with_operation_total",
		Help: "Total number of logs generated with a given Operation",
	}, []string{"Operation"})

totalAlertsWithAction = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_relay_logs_with_action_total",
		Help: "Total number of logs generated with a given action",
	}, []string{"Action"})
)

func init() {
	prometheus.MustRegister(totalAlertsRequestsinHost)
	prometheus.MustRegister(totalAlertsRequestsinPod)
	prometheus.MustRegister(totalAlertsRequestsinNamespace)
	prometheus.MustRegister(totalAlertsRequestsinContainer)
	prometheus.MustRegister(totalAlertsWithPolicy)
	prometheus.MustRegister(totalAlertsWithSeverity)
	prometheus.MustRegister(totalAlertsWithType)
	prometheus.MustRegister(totalAlertsWithOperation)
	prometheus.MustRegister(totalAlertsWithAction)
}


func GetPrometheusAlerts(wg *sync.WaitGroup) {
	url := "kubearmor.kube-system.svc.cluster.local"
	port := "32767"
	address := url + ":" + port

	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
	}

	client := pb.NewLogServiceClient(connection)

	req := &pb.RequestMessage{
		Filter: "policy",
	}

	stream, err := client.WatchAlerts(context.Background(), req)
	if err != nil {
		fmt.Printf("Failed to call WatchAlerts() (%s)\n", err.Error())
	}

	for {
		alertIn, err := stream.Recv()
		if err != nil {
			fmt.Printf("Failed to receive any alerts (%s)\n", err.Error())
			break
		}

		switch err {
		case io.EOF:
			fmt.Println(err.Error())
			break
		case nil:
		default:
			fmt.Println(err.Error())
		}

		fmt.Println(alertIn)
		totalAlertsRequestsinHost.WithLabelValues(alertIn.HostName).Add(1)
		totalAlertsRequestsinPod.WithLabelValues(alertIn.PodName).Add(1)
		totalAlertsRequestsinNamespace.WithLabelValues(alertIn.NamespaceName).Add(1)
		totalAlertsRequestsinContainer.WithLabelValues(alertIn.ContainerName).Add(1)
		totalAlertsWithPolicy.WithLabelValues(alertIn.PolicyName).Add(1)
		totalAlertsWithSeverity.WithLabelValues(alertIn.Severity).Add(1)
		totalAlertsWithType.WithLabelValues(alertIn.Type).Add(1)
		totalAlertsWithOperation.WithLabelValues(alertIn.Operation).Add(1)
		totalAlertsWithAction.WithLabelValues(alertIn.Action).Add(1)
	}
	wg.Done()
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go GetPrometheusAlerts(&wg)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9100", nil))
	wg.Wait()
}
