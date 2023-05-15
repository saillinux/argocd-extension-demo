package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type InstanceTemplate struct {
	Name        string   `json:"name"`
	MachineType string   `json:"machineType"`
	Network     string   `json:"network"`
	SourceImage string   `json:"sourceImage"`
	Subnetwork  string   `json:"subnetwork"`
	Tags        []string `json:"tags"`
	TargetSize  int64    `json:"targetSize"`
}

type ManagedInstance struct {
	Name             string `json:"instance"`
	InstanceTemplate string `json:"instanceTemplate"`
	Zone             string `json:"zone"`
	Status           string `json:"status"`
	SelfLink         string `json:"selfLink"`
}

type ManagedInstanceGroup struct {
	ProjectId        string              `json:"projectId"`
	Region           string              `json:"region"`
	Name             string              `json:"groupName"`
	TargetSize       int64               `json:"targetSize"`
	Versions         []*InstanceTemplate `json:"versions"`
	BaseInstanceName string              `json:"baseInstanceName"`
	ManagedInstances []*ManagedInstance  `json:"managedInstances"`
	UpdatePolicy     string              `json:"updatePolicy"`
	Status           bool                `json:"status"`
	SelfLink         string              `json:"selfLink"`
}

func getManagedInstanceGroupHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len("/compute/instancegroup/get/"):]
	parts := strings.Split(target, "/")
	projectId, region, instanceGroupName := parts[0], parts[1], parts[2]

	if projectId == "" || region == "" || instanceGroupName == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}

	log.Printf("projectId: %s, region: %s, instanceGroupName: %s", projectId, region, instanceGroupName)

	ctx := context.Background()

	client, err := compute.NewService(ctx, option.WithScopes(compute.ComputeScope))
	if err != nil {
		log.Printf("Failed to create Compute Engine API client: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Call the Compute Engine API to get the managed instance group information
	instanceGroup, err := client.RegionInstanceGroupManagers.Get(projectId, region, instanceGroupName).Context(ctx).Do()
	if err != nil {
		// log.Fatalf("Failed to retrieve managed instance group: %v", err)
		log.Printf("Failed to retrieve managed instance group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	instanceList, err := client.RegionInstanceGroupManagers.ListManagedInstances(projectId, region, instanceGroupName).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	var versions []*InstanceTemplate

	if instanceGroup.InstanceTemplate != "" {
		tmplParts := strings.Split(instanceGroup.InstanceTemplate, "/")
		instanceTemplateName := tmplParts[len(tmplParts)-1]
		instanceTemplate, err := client.InstanceTemplates.Get(projectId, instanceTemplateName).Context(ctx).Do()
		if err != nil {
			log.Print(err)
		}

		it := &InstanceTemplate{
			Name:        instanceTemplate.Name,
			MachineType: instanceTemplate.Properties.MachineType,
			Network:     instanceTemplate.Properties.NetworkInterfaces[0].Network,
			SourceImage: instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage,
			Subnetwork:  instanceTemplate.Properties.NetworkInterfaces[0].Subnetwork,
			Tags:        instanceTemplate.Properties.Tags.Items,
			TargetSize:  0,
		}

		versions = append(versions, it)
	} else {
		for _, version := range instanceGroup.Versions {
			tmplParts := strings.Split(version.InstanceTemplate, "/")
			instanceTemplateName := tmplParts[len(tmplParts)-1]
			instanceTemplate, err := client.InstanceTemplates.Get(projectId, instanceTemplateName).Context(ctx).Do()
			if err != nil {
				log.Print(err)
			}

			it := &InstanceTemplate{
				Name:        instanceTemplate.Name,
				MachineType: instanceTemplate.Properties.MachineType,
				Network:     instanceTemplate.Properties.NetworkInterfaces[0].Network,
				SourceImage: instanceTemplate.Properties.Disks[0].InitializeParams.SourceImage,
				Subnetwork:  instanceTemplate.Properties.NetworkInterfaces[0].Subnetwork,
				Tags:        instanceTemplate.Properties.Tags.Items,
				TargetSize:  version.TargetSize.Fixed,
			}

			versions = append(versions, it)
		}
	}

	mig := &ManagedInstanceGroup{
		ProjectId:        projectId,
		Region:           region,
		Name:             instanceGroupName,
		TargetSize:       instanceGroup.TargetSize,
		Versions:         versions,
		BaseInstanceName: instanceGroup.BaseInstanceName,
		UpdatePolicy:     instanceGroup.UpdatePolicy.Type,
		Status:           instanceGroup.Status.IsStable,
		SelfLink:         instanceGroup.SelfLink,
		ManagedInstances: []*ManagedInstance{},
	}

	for _, instance := range instanceList.ManagedInstances {
		parts := strings.Split(instance.Instance, "/")
		zone, instanceName := parts[8], parts[10]
		// when an instance is being deleted the reference value to the version is no longer exist and causes invalid memory access
		if instance.Version == nil {
			continue
		}
		tmplParts := strings.Split(instance.Version.InstanceTemplate, "/")
		instanceTemplateName := tmplParts[len(tmplParts)-1]
		mig.ManagedInstances = append(mig.ManagedInstances, &ManagedInstance{
			Name:             instanceName,
			InstanceTemplate: instanceTemplateName,
			Zone:             zone,
			Status:           instance.InstanceStatus,
			SelfLink:         instance.Instance,
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

func updateManagedInstanceGroupHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len("/compute/instancegroup/update/"):]
	parts := strings.Split(target, "/")
	projectId, region, instanceGroupName := parts[0], parts[1], parts[2]
	if projectId == "" || region == "" || instanceGroupName == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}
	log.Printf("projectId: %s, region: %s, instanceGroupName: %s", projectId, region, instanceGroupName)

	strategy := r.URL.Query().Get("strategy")
	if strategy == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}

	targetTemplate := r.URL.Query().Get("target_template")
	if targetTemplate == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}

	targetSizeStr := r.URL.Query().Get("target_size")
	if targetSizeStr == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}

	targetSize, err := strconv.ParseInt(targetSizeStr, 10, 64)

	if err != nil {
		fmt.Println("Error during conversion")
		return
	}

	log.Printf("Deployment %s %s %d", strategy, targetTemplate, targetSize)

	ctx := context.Background()

	client, err := compute.NewService(ctx, option.WithScopes(compute.ComputeScope))
	if err != nil {
		log.Printf("Failed to create Compute Engine API client: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	instanceTemplate, err := client.InstanceTemplates.Get(projectId, targetTemplate).Context(ctx).Do()
	if err != nil {
		log.Printf("Failed to retrieve the instance teamplate: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Target instance template: %s", instanceTemplate.SelfLink)

	instanceGroup, err := client.RegionInstanceGroupManagers.Get(projectId, region, instanceGroupName).Context(ctx).Do()
	if err != nil {
		// log.Fatalf("Failed to retrieve managed instance group: %v", err)
		log.Printf("Failed to retrieve managed instance group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Current instance template: %s", instanceGroup.InstanceTemplate)

	var patchRequest *compute.InstanceGroupManager = nil

	currentTemplateLink := instanceGroup.InstanceTemplate
	targetTemplateSelfLink := instanceTemplate.SelfLink

	if strategy == "rolling" {
		// Create the instance group manager patch request with the new instance template
		patchRequest = &compute.InstanceGroupManager{
			InstanceTemplate: targetTemplateSelfLink,
			UpdatePolicy: &compute.InstanceGroupManagerUpdatePolicy{
				Type:          "PROACTIVE",
				MinimalAction: "REPLACE",
				MaxUnavailable: &compute.FixedOrPercent{
					Fixed: 0,
				},
			},
		}
	} else if strategy == "canary" {
		log.Printf("Canary Update using %s", targetTemplate)
		// Create the instance group manager patch request with the new instance template
		patchRequest = &compute.InstanceGroupManager{
			Versions: []*compute.InstanceGroupManagerVersion{
				{
					TargetSize: &compute.FixedOrPercent{
						Fixed: targetSize,
					},
					InstanceTemplate: targetTemplateSelfLink,
				},
				{
					InstanceTemplate: currentTemplateLink,
				},
			},
		}
	} else {
		log.Printf("Unknown strategy: %s", strategy)
		http.Error(w, "Unknown strategy", http.StatusBadRequest)
		return
	}

	// Make the PATCH request to update the instance group manager
	_, err = client.RegionInstanceGroupManagers.Patch(projectId, region, instanceGroupName, patchRequest).Context(ctx).Do()
	if err != nil {
		log.Printf("Failed to update instance group manager: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Print("Instance group manager updated successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}

func listInstanceTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len("/compute/instancetemplate/list/"):]
	parts := strings.Split(target, "/")
	projectId := parts[0]
	if projectId == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}
	log.Printf("projectId: %s", projectId)

	ctx := context.Background()

	client, err := compute.NewService(ctx, option.WithScopes(compute.ComputeScope))
	if err != nil {
		log.Printf("Failed to create Compute Engine API client: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	instanceTemplates, err := client.InstanceTemplates.List(projectId).Context(ctx).Do()
	if err != nil {
		log.Printf("Failed to retrieve instance templates: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var instanceTemplateNames []string
	for _, instanceTemplate := range instanceTemplates.Items {
		instanceTemplateNames = append(instanceTemplateNames, instanceTemplate.Name)
	}

	instanceTemplateNamesJson, err := json.Marshal(instanceTemplateNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(instanceTemplateNamesJson)
}

func main() {
	http.HandleFunc("/compute/instancegroup/get/", getManagedInstanceGroupHandler)
	http.HandleFunc("/compute/instancegroup/update/", updateManagedInstanceGroupHandler)
	http.HandleFunc("/compute/instancetemplate/list/", listInstanceTemplatesHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
