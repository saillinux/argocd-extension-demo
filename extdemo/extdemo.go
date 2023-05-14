package main

import (
	"context"
	"encoding/json"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"strings"
)

type InstanceTemplate struct {
	Name        string   `json:"name"`
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
	// TargetPools
	// StatefulPolicy
	// Versions
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

	// client.RegionInstanceGroupManagers.Patch(projectId, region, instanceGroupName, &compute.RegionInstanceGroupManager{}).Context(ctx).Do()

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
			Subnetwork:  instanceTemplate.Properties.NetworkInterfaces[0].Subnetwork,
			Tags:        instanceTemplate.Properties.Tags.Items,
		}

		versions = append(versions, it)
	} else {
		for _, version := range instanceGroup.Versions {
			log.Printf("Version: %s, %s, %d", version.Name, version.InstanceTemplate, version.TargetSize.Fixed)
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
				Subnetwork:  instanceTemplate.Properties.NetworkInterfaces[0].Subnetwork,
				Tags:        instanceTemplate.Properties.Tags.Items,
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
	log.Print(strategy)

	targetTemplate := r.URL.Query().Get("target_template")
	if targetTemplate == "" {
		http.Error(w, "Missing parameter(s)", http.StatusBadRequest)
		return
	}
	log.Print(targetTemplate)

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

	log.Printf("Current: %s", instanceTemplate.SelfLink)

	var patchRequest *compute.InstanceGroupManager = nil

	splitSelfLink := strings.Split(instanceTemplate.SelfLink, "/")
	currentTemplate := splitSelfLink[len(splitSelfLink)-1]

	newInstanceTemplateSelfLink := strings.Replace(instanceTemplate.SelfLink, currentTemplate, targetTemplate, 1)
	log.Printf("New: %s", newInstanceTemplateSelfLink)

	if strategy == "rolling" {
		// Create the instance group manager patch request with the new instance template
		patchRequest = &compute.InstanceGroupManager{
			InstanceTemplate: newInstanceTemplateSelfLink,
			UpdatePolicy: &compute.InstanceGroupManagerUpdatePolicy{
				Type: "PROACTIVE",
			},
		}
	} else if strategy == "canary" {
		log.Printf("Canary Update using %s", targetTemplate)
		// Create the instance group manager patch request with the new instance template
		patchRequest = &compute.InstanceGroupManager{
			Versions: []*compute.InstanceGroupManagerVersion{
				{
					TargetSize: &compute.FixedOrPercent{
						Fixed: 1,
					},
					InstanceTemplate: newInstanceTemplateSelfLink,
				},
				{
					InstanceTemplate: instanceTemplate.SelfLink,
				},
			},
		}
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
