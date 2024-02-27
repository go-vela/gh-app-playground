package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"
)

func main() {
	tr := http.DefaultTransport

	appID, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		panic(err)
	}

	appPrivateKeyFile := os.Getenv("APP_PRIVATE_KEY_FILE")
	gitApiURL := os.Getenv("APP_GIT_API_URL")
	targetOrg := os.Getenv("APP_TARGET_ORG")
	targetRepo := os.Getenv("APP_TARGET_REPO")

	// Wrap the shared transport for use with the app ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.NewAppsTransportKeyFromFile(tr, appID, appPrivateKeyFile)
	if err != nil {
		panic(err)
	}

	itr.BaseURL = gitApiURL
	client, err := github.NewClient(&http.Client{Transport: itr}).WithEnterpriseURLs(gitApiURL, gitApiURL)
	if err != nil {
		panic(err)
	}

	installID := int64(11602)

	token, _, err := client.Apps.CreateInstallationToken(context.Background(), installID, &github.InstallationTokenOptions{})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.GetToken()},
	)
	tc := oauth2.NewClient(ctx, ts)

	secondClient, err := github.NewClient(tc).WithEnterpriseURLs(gitApiURL, gitApiURL)
	if err != nil {
		panic(err)
	}

	// Create a new check suite
	checkCreateOpts := github.CreateCheckRunOptions{
		Name:    "vela-testing",
		HeadSHA: "main",
	}

	content, _, err := secondClient.Checks.CreateCheckRun(context.Background(), targetOrg, targetRepo, checkCreateOpts)
	if err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)

	checkRunOpts := github.UpdateCheckRunOptions{
		Name:       "vela-testing",
		Status:     github.String("completed"),
		Conclusion: github.String("success"),
	}

	contentResp, resp, err := secondClient.Checks.UpdateCheckRun(context.Background(), targetOrg, targetRepo, content.GetID(), checkRunOpts)

	if err != nil {
		panic(err)
	}

	fmt.Println(contentResp)
	fmt.Println(resp)
}
