package main

import (
	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	// "google.golang.org/api/compute/v1"
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

	client, err := compute.NewInstanceGroupManagersRESTClient(ctx)
	if err != nil {
		log.Printf("NewInstancesRESTClient: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	/* 	client, err := compute.NewRegionInstanceGroupManagersRESTClient(ctx)
	   	if err != nil {
	   		log.Printf("NewInstancesRESTClient: %v", err)
	   		http.Error(w, err.Error(), http.StatusInternalServerError)
	   		return
	   	}
	   	defer client.Close() */

	req := &computepb.GetInstanceGroupManagerRequest{
		Project:              projectId,
		Zone:                 "us-central1-c",
		InstanceGroupManager: "gke-front01-us-central1--default-pool-03af78fd-grp",
	}

	/* 	req := &computepb.GetRegionInstanceGroupManagerRequest{
		Project:              projectId,
		Region:               region,
		InstanceGroupManager: instanceGroupName,
	} */

	resp, err := client.Get(ctx, req)
	if err != nil {
		log.Printf("unable to retrieve instanceGroup: %v", err)
		return
	}

	log.Printf("%+v\n", resp)

	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte("success"))
}

func main() {
	http.HandleFunc("/storage/list", listBucketsHandler)
	http.HandleFunc("/compute/instancegroup/", getManagedInstanceGroupHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
