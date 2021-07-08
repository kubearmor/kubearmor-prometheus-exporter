package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"io"
	"log"
	"net/http"
	"sync"

	pb "github.com/kubearmor/KubeArmor/protobuf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
totalAlertsRequestsinHost = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_in_host_total",
		Help: "Total number of alerts based on HostName",
	}, []string{"HostName"})

totalAlertsRequestsinNamespace = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_in_namespace_total",
		Help: "Total number of alerts based on Namespace",
	}, []string{"NamespaceName"})

totalAlertsRequestsinPod = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_in_pod_total",
		Help: "Total number of alerts based on PodName",
	}, []string{"PodName"})

totalAlertsRequestsinContainer = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_in_container_total",
		Help: "Total number of alerts based on Container",
	}, []string{"ContainerName"})

totalAlertsWithPolicy = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_with_policy_total",
		Help: "Total number of alerts based on Policy",
	}, []string{"PolicyName"})

totalAlertsWithSeverity = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_with_severity_total",
		Help: "Total number of alerts with X severity or above",
	}, []string{"Severity"})

totalAlertsWithType = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_with_type_total",
		Help: "Total number of alerts based on Type",
	}, []string{"Type"})

totalAlertsWithOperation = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_with_operation_total",
		Help: "Total number of alerts based on Operation",
	}, []string{"Operation"})

totalAlertsWithAction = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kubearmor_alerts_with_action_total",
		Help: "Total number of alerts based on Action",
	}, []string{"Action"})
)

func init() {
	prometheus.MustRegister(totalAlertsRequestsinHost)
	prometheus.MustRegister(totalAlertsRequestsinNamespace)
	prometheus.MustRegister(totalAlertsRequestsinPod)
	prometheus.MustRegister(totalAlertsRequestsinContainer)

	prometheus.MustRegister(totalAlertsWithPolicy)
	prometheus.MustRegister(totalAlertsWithSeverity)
	prometheus.MustRegister(totalAlertsWithType)
	prometheus.MustRegister(totalAlertsWithOperation)
	prometheus.MustRegister(totalAlertsWithAction)
}


func GetPrometheusAlerts(wg *sync.WaitGroup, gRPCAddr string) {
	connection, err := grpc.Dial(gRPCAddr, grpc.WithInsecure())
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
			//
		default:
			fmt.Println(err.Error())
		}

		// fmt.Println(alertIn)

		totalAlertsRequestsinHost.WithLabelValues(alertIn.HostName).Add(1)
		totalAlertsRequestsinNamespace.WithLabelValues(alertIn.NamespaceName).Add(1)
		totalAlertsRequestsinPod.WithLabelValues(alertIn.PodName).Add(1)
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

	// == //
	
	gRPCPtr := flag.String("gRPC", "", "gRPC server information")
	flag.Parse()
	
	// == //
	
	gRPCAddr := ""
	
	if *gRPCPtr != "" {
		gRPCAddr = *gRPCPtr
	} else {
		if val, ok := os.LookupEnv("KUBEARMOR_SERVICE"); ok {
			gRPCAddr = val
		} else {
			gRPCAddr = "localhost:32767"
		}
	}
	
	// == //
	
	wg.Add(1)
	go GetPrometheusAlerts(&wg, gRPCAddr)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9100", nil))

	wg.Wait()
}
