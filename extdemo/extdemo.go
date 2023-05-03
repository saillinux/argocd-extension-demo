package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
)

type Buckets struct {
	Buckets []string `json:"buckets"`
}

type InstanceTemplate struct {
	Name        string   `json:"name"`
	Region      string   `json:"region"`
	MachineType string   `json:"machineType"`
	Network     string   `json:"network"`
	Subnetwork  string   `json:"subnetwork"`
	Tags        []string `json:"tags"`
}

type ManagedInstance struct {
	Name     string `json:"instance"`
	Zone     string `json:"zone"`
	Status   string `json:"status"`
	SelfLink string `json:"selfLink"`
}

type ManagedInstanceGroup struct {
	ProjectId        string             `json:"projectId"`
	Region           string             `json:"region"`
	Name             string             `json:"groupName"`
	TargetSize       int64              `json:"targetSize"`
	InstanceTemplate *InstanceTemplate  `json:"instanceTemplate"`
	BaseInstanceName string             `json:"baseInstanceName"`
	ManagedInstances []*ManagedInstance `json:"managedInstances"`
	UpdatePolicy     string             `json:"updatePolicy"`
	Status           bool               `json:"status"`
	SelfLink         string             `json:"selfLink"`
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

	// TODO, do we need this?
	_, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Print(err)
	}

	client, err := compute.NewService(ctx)
	if err != nil {
		log.Printf("Failed to create Compute Engine API client: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Call the Compute Engine API to get the managed instance group information
	instanceGroup, err := client.RegionInstanceGroupManagers.Get(projectId, region, instanceGroupName).Context(ctx).Do()
	// instanceGroup, err := client.InstanceGroupManagers.Get(projectId, region+"-b", instanceGroupName).Context(ctx).Do()
	if err != nil {
		// log.Fatalf("Failed to retrieve managed instance group: %v", err)
		log.Printf("Failed to retrieve managed instance group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Call the Compute Engine API to get the instances in the instance group
	// TODO, replace it with computeService.RegionInstanceGroupManagers.ListManagedInstances(project, region, instanceGroupManager).Context(ctx).Do()
	instanceList, err := client.RegionInstanceGroups.ListInstances(
		projectId,
		region,
		instanceGroupName,
		&compute.RegionInstanceGroupsListInstancesRequest{},
	).Do()
	if err != nil {
		log.Printf("Failed to retrieve instances from instance group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplParts := strings.Split(instanceGroup.InstanceTemplate, "/")
	instanceTemplateName := tmplParts[len(tmplParts)-1]
	instanceTemplate, err := client.InstanceTemplates.Get(projectId, instanceTemplateName).Context(ctx).Do()
	if err != nil {
		log.Print(err)
	}

	it := &InstanceTemplate{
		Name:        instanceTemplate.Name,
		Region:      instanceTemplate.Region,
		MachineType: instanceTemplate.Properties.MachineType,
		Network:     instanceTemplate.Properties.NetworkInterfaces[0].Network,
		Subnetwork:  instanceTemplate.Properties.NetworkInterfaces[0].Subnetwork,
		Tags:        instanceTemplate.Properties.Tags.Items,
	}

	mig := &ManagedInstanceGroup{
		ProjectId:        projectId,
		Region:           region,
		Name:             instanceGroupName,
		TargetSize:       instanceGroup.TargetSize,
		InstanceTemplate: it,
		BaseInstanceName: instanceGroup.BaseInstanceName,
		UpdatePolicy:     instanceGroup.UpdatePolicy.Type,
		Status:           instanceGroup.Status.IsStable,
		SelfLink:         instanceGroup.SelfLink,
		ManagedInstances: []*ManagedInstance{},
	}

	for _, instance := range instanceList.Items {
		parts := strings.Split(instance.Instance, "/")
		zone, instanceName := parts[8], parts[10]

		mig.ManagedInstances = append(mig.ManagedInstances, &ManagedInstance{
			Name:     instanceName,
			Zone:     zone,
			Status:   instance.Status,
			SelfLink: instance.Instance,
		})
	}

	migJson, err := json.Marshal(mig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(migJson)
}

func main() {
	http.HandleFunc("/storage/list", listBucketsHandler)
	http.HandleFunc("/compute/instancegroup/", getManagedInstanceGroupHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
