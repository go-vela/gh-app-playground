package main

import (
	"context"
	"fmt"
	"math/rand"
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

	fmt.Printf("Using App ID: %d\n", appID)

	appPrivateKeyFile := os.Getenv("APP_PRIVATE_KEY_FILE")
	gitApiURL := os.Getenv("APP_GIT_API_URL")
	targetOrg := os.Getenv("APP_TARGET_ORG")
	targetRepo := os.Getenv("APP_TARGET_REPO")

	fmt.Printf("Sourcing private key from path: %s\n", appPrivateKeyFile)

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

	installations, _, err := client.Apps.ListInstallations(context.Background(), &github.ListOptions{})
	if err != nil {
		panic(err)
	}

	var installID int64
	for _, install := range installations {
		installID = install.GetID()
		fmt.Println(install)
	}

	fmt.Println(installID)

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

	var checkRuns []*github.CheckRun

	for i := 0; i < 5; i++ {
		// Create a new check suite
		checkCreateOpts := github.CreateCheckRunOptions{
			Name:    fmt.Sprintf("vela-testing-%d", i),
			HeadSHA: "main",
		}

		content, _, err := secondClient.Checks.CreateCheckRun(context.Background(), targetOrg, targetRepo, checkCreateOpts)
		if err != nil {
			panic(err)
		}

		checkRuns = append(checkRuns, content)
	}

	for _, checkRun := range checkRuns {
		time.Sleep(time.Duration(rand.Intn(6)+5) * time.Second)

		checkRunOpts := github.UpdateCheckRunOptions{
			Name:       checkRun.GetName(),
			Status:     github.String("completed"),
			Conclusion: github.String("success"),
		}

		_, _, err := secondClient.Checks.UpdateCheckRun(context.Background(), targetOrg, targetRepo, checkRun.GetID(), checkRunOpts)

		if err != nil {
			panic(err)
		}
	}
}
