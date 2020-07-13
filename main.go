package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/exoscale/egoscale"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

type StaticSDConfig []TargetGroup

func getInstancePoolInstanceIps(client *egoscale.Client, zoneId *egoscale.UUID, poolId *egoscale.UUID) ([]string, error) {
	ctx := context.Background()
	resp, err := client.RequestWithContext(ctx, egoscale.GetInstancePool{
		ZoneID: zoneId,
		ID:     poolId,
	})
	if err != nil {
		log.Printf("failed to get instance pool from Exoscale (%v)", err)
		// Ignore error. next run will hopefully work better
		return []string{}, err
	}
	response := resp.(*egoscale.GetInstancePoolResponse)
	if len(response.InstancePools) == 0 {
		log.Fatalf("instance pool not found")
	} else if len(response.InstancePools) > 1 {
		//This should never happen
		log.Fatalf("more than one instance pool returned")
	}
	instancePool := response.InstancePools[0]

	var ips []string
	for _, vm := range instancePool.VirtualMachines {
		ips = append(ips, vm.Nic[0].IPAddress.String())
	}

	return ips, nil
}

func main() {
	instancePoolId := ""
	exoscaleEndpoint := "https://api.exoscale.ch/v1/"
	exoscaleZoneId := ""
	exoscaleApiKey := ""
	exoscaleApiSecret := ""
	prometheusFile := "/var/run/prometheus-sd-exoscale-instance-pools/config.json"
	flag.StringVar(
		&instancePoolId,
		"instance-pool-id",
		instancePoolId,
		"ID of the instance pool to query",
	)
	flag.StringVar(
		&exoscaleZoneId,
		"exoscale-zone-id",
		exoscaleZoneId,
		"Exoscale zone ID",
	)
	flag.StringVar(
		&exoscaleEndpoint,
		"exoscale-endpoint",
		exoscaleEndpoint,
		"Endpoint URL of the Exoscale API",
	)
	flag.StringVar(
		&exoscaleApiKey,
		"exoscale-api-key",
		exoscaleApiKey,
		"API key for Exoscale",
	)
	flag.StringVar(
		&exoscaleApiSecret,
		"exoscale-api-secret",
		exoscaleApiSecret,
		"API secret for Exoscale",
	)
	flag.StringVar(
		&prometheusFile,
		"prometheus-file",
		prometheusFile,
		"Static service discovery file for Prometheus",
	)
	flag.Parse()

	zoneId, err := egoscale.ParseUUID(exoscaleZoneId)
	if err != nil {
		log.Fatalf("invalid zone ID (%v)", err)
	}

	poolId, err := egoscale.ParseUUID(instancePoolId)
	if err != nil {
		log.Fatalf("invalid pool ID (%v)", err)
	}

	exoscaleClient := egoscale.NewClient(exoscaleEndpoint, exoscaleApiKey, exoscaleApiSecret)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		ips, err := getInstancePoolInstanceIps(exoscaleClient, zoneId, poolId)
		if err != nil {
			log.Fatalf("failed to fetch instance pool IPs, crashing out (%v)", err)
		}

		config := StaticSDConfig{TargetGroup{
			Targets: ips,
			Labels:  map[string]string{},
		}}

		file, err := json.MarshalIndent(config, "", " ")
		if err != nil {
			log.Fatalf("failed to create JSON (%v)", err)
		}

		err = ioutil.WriteFile(prometheusFile, file, 0644)
		if err != nil {
			log.Fatalf("failed to write Prometheus file %s (%v)", prometheusFile, err)
		}

		select {
		case _, _ = <-sigs:
			fmt.Println("interrupt received, shutting down")
			break
		default:
			time.Sleep(10 * time.Second)
		}
	}
}
