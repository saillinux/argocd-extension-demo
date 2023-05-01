package main

import (
	// compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
	"log"
	"net/http"
	"strings"
	"time"
)

type Buckets struct {
	Buckets []string `json:"buckets"`
}

type ManagedInstance struct {
	ProjectId string `json:"projectId"`
	Zone      string `json:"zone"`
	Name      string `json:"instance"`
	Status    string `json:"status"`
	SelfLink  string `json:"selfLink"`
}

type ManagedInstanceGroup struct {
	ProjectId             string            `json:"projectId"`
	Region                string            `json:"region"`
	Name                  string            `json:"groupName"`
	TargetSize            int64             `json:"targetSize"`
	InstanceGroupTemplate string            `json:"instanceGroupTemplate"`
	BaseInstanceName      string            `json:"baseInstanceName"`
	ManagedInstances      []ManagedInstance `json:"managedInstances"`
	UpdatePolicy          string            `json:"updatePolicy"`
	Status                bool              `json:"status"`
	SelfLink              string            `json:"selfLink"`
	// TargetPools
	// StatefulPolicy
	// Versions
}

// listBuckets lists buckets in the project.
func listBuckets(projectID string) ([]string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var buckets []string
	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, battrs.Name)
	}
	return buckets, nil
}

func listBucketsHandler(w http.ResponseWriter, r *http.Request) {
	PROJECT_ID := "heewonk-bunker"
	buckets, err := listBuckets(PROJECT_ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bucketsJson, err := json.Marshal(buckets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(buckets)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bucketsJson)
}

func getManagedInstanceGroupHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len("/compute/instancegroup/"):]
	parts := strings.Split(target, "/")
	projectId, region, instanceGroupName := parts[0], parts[1], parts[2]

	if projectId == "" || region == "" || instanceGroupName == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}

	log.Printf("projectId: %s, region: %s, instanceGroupName: %s", projectId, region, instanceGroupName)

	ctx := context.Background()
	client, err := compute.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed to create Compute Engine API client: %v", err)
	}

	// Call the Compute Engine API to get the managed instance group information
	instanceGroup, err := client.RegionInstanceGroupManagers.Get(projectId, region, instanceGroupName).Do()
	if err != nil {
		log.Fatalf("Failed to retrieve managed instance group: %v", err)
	}

	// fmt.Printf("%+v\n", instanceGroup)

	// Print out the name and size of the managed instance group
	/* 	fmt.Printf("Managed instance group id: %d\n", instanceGroup.Id)
	   	fmt.Printf("Managed instance group name: %s\n", instanceGroup.Name)
	   	fmt.Printf("Managed instance group: %s\n", instanceGroup.InstanceGroup)
	   	fmt.Printf("Managed instance group target size: %d\n", instanceGroup.TargetSize)
	   	fmt.Printf("Managed instance group template: %s\n", instanceGroup.InstanceTemplate) */
	/* 	fmt.Println(instanceGroup.Status.IsStable)
	   	fmt.Println(len(instanceGroup.Versions))
	   	fmt.Println(instanceGroup.UpdatePolicy.Type) */

	// Call the Compute Engine API to get the instances in the instance group
	instanceList, err := client.RegionInstanceGroups.ListInstances(
		projectId,
		region,
		instanceGroupName,
		&compute.RegionInstanceGroupsListInstancesRequest{},
	).Do()
	if err != nil {
		log.Fatalf("Failed to retrieve instances from instance group: %v", err)
	}

	tmplParts := strings.Split(instanceGroup.InstanceTemplate, "/")

	mig := &ManagedInstanceGroup{
		ProjectId:             projectId,
		Region:                region,
		Name:                  instanceGroupName,
		TargetSize:            instanceGroup.TargetSize,
		InstanceGroupTemplate: tmplParts[len(tmplParts)-1],
		BaseInstanceName:      instanceGroup.BaseInstanceName,
		UpdatePolicy:          instanceGroup.UpdatePolicy.Type,
		Status:                instanceGroup.Status.IsStable,
		SelfLink:              instanceGroup.SelfLink,
		ManagedInstances:      []ManagedInstance{},
	}

	for _, instance := range instanceList.Items {
		parts := strings.Split(instance.Instance, "/")
		projectId, zone, instanceName := parts[6], parts[8], parts[10]

		/* 		fmt.Println("=====================================")
		   		fmt.Printf("Instance ProjectId: %s\n", projectId)
		   		fmt.Printf("Instance Zone: %s\n", zone)
		   		fmt.Printf("Instance Name: %s\n", instanceName)
		   		fmt.Printf("Instance Status: %s\n", instance.Status)
		   		fmt.Printf("Instance Resource URI: %s\n", instance.Instance) */

		mig.ManagedInstances = append(mig.ManagedInstances, ManagedInstance{
			ProjectId: projectId,
			Zone:      zone,
			Name:      instanceName,
			Status:    instance.Status,
			SelfLink:  instance.Instance,
		})
	}

	migJson, err := json.Marshal(mig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Println(mig)
	w.Header().Set("Content-Type", "application/json")
	w.Write(migJson)
}

func main() {
	http.HandleFunc("/storage/list", listBucketsHandler)
	http.HandleFunc("/compute/instancegroup/", getManagedInstanceGroupHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
