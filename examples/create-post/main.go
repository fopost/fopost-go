// Command create-post composes a post, previews it, and optionally publishes.
//
//	export FOPOST_API_KEY=fp_...
//	go run ./examples/create-post "Hello from the Go SDK" --publish
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fopost/fopost-go"
)

func main() {
	args := os.Args[1:]
	publish := false
	var words []string
	for _, arg := range args {
		if arg == "--publish" {
			publish = true
			continue
		}
		words = append(words, arg)
	}
	text := strings.Join(words, " ")
	if text == "" {
		text = "Hello from the FoPost Go SDK"
	}

	client, err := fopost.New("") // reads FOPOST_API_KEY
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	workspaces, err := client.Workspaces.List(ctx)
	if err != nil {
		log.Fatalf("listing workspaces: %v", err)
	}
	if len(workspaces) == 0 {
		log.Fatal("no workspaces on this account")
	}
	workspace := workspaces[0]

	accounts, err := client.Accounts.List(ctx, workspace.ID)
	if err != nil {
		log.Fatalf("listing accounts: %v", err)
	}
	if len(accounts) == 0 {
		log.Fatalf("no connected accounts in %s", workspace.Name)
	}

	ids := make([]string, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}

	post, err := client.Posts.Create(ctx, &fopost.CreatePostRequest{
		WorkspaceID: workspace.ID,
		Accounts:    ids,
		Content:     fopost.Text(text),
	})
	if err != nil {
		log.Fatalf("creating the post: %v", err)
	}
	fmt.Printf("created %s in %s (%s)\n", post.ID, workspace.Name, post.Status)

	preflight, err := client.Posts.Preflight(ctx, post.ID)
	if err != nil {
		log.Fatalf("preflight: %v", err)
	}
	for _, account := range preflight.Accounts {
		fmt.Printf("  %-12s ready=%v issues=%v\n", account.Platform, account.Ready, account.Issues)
	}

	if !publish {
		fmt.Println("draft saved — pass --publish to send it")
		return
	}

	result, err := client.Posts.Publish(ctx, post.ID, nil)
	if err != nil {
		if apiErr, ok := fopost.APIError(err); ok && fopost.IsPaymentRequired(err) {
			log.Fatalf("publishing needs an active subscription: %s", apiErr.UpgradeURL())
		}
		log.Fatalf("publishing: %v", err)
	}
	fmt.Printf("post is %s, %d deliveries queued\n", result.PostStatus, len(result.Deliveries))
}
