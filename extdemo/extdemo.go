package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/iterator"
	"log"
	"net/http"
	"time"
)

var PROJECT_ID = "heewonk-bunker"

type Buckets struct {
	Buckets []string `json:"buckets"`
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

func main() {
	http.HandleFunc("/storage/list", listBucketsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
