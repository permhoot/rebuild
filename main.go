// Copyright © 2024 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	shpbuildAPI "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	shpbuild "github.com/shipwright-io/build/pkg/client/clientset/versioned/typed/build/v1beta1"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	serving "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

const (
	serverAddrPort               = 8080
	defaultBuildRunWaitTimeout   = time.Duration(10 * time.Minute)
	updateTimestampAnnotationKey = "client.knative.dev/updateTimestamp"
	namespaceFile                = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	userImageAnnotation          = "client.knative.dev/user-image"
)

func main() {
	ctx := context.Background()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	srv := &http.Server{Addr: fmt.Sprintf(":%d", serverAddrPort)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var decoder = json.NewDecoder(r.Body)

			var obj event
			if err := decoder.Decode(&obj); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var repo = strings.TrimSuffix(obj.Repository.CloneURL, ".git")
			log.Printf("checking for build or build-run that covers %s\n", repo)
			w.WriteHeader(handleRepo(ctx, &wg, repo))

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-signals
	log.Printf("received stop signal, waiting for all pending work to finish")
	wg.Wait()

	log.Printf("shutting down server")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server: %v", err)
	}
}

func inClusterSetup() (string, *shpbuild.ShipwrightV1beta1Client, *serving.ServingV1Client, error) {
	out, err := os.ReadFile(namespaceFile)
	if err != nil {
		return "", nil, nil, err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return "", nil, nil, err
	}

	buildClient, err := shpbuild.NewForConfig(config)
	if err != nil {
		return "", nil, nil, err
	}

	servingClient, err := serving.NewForConfig(config)
	if err != nil {
		return "", nil, nil, err
	}

	return string(out), buildClient, servingClient, nil
}

func findBuild(ctx context.Context, buildClient *shpbuild.ShipwrightV1beta1Client, namespace string, repo string) (*shpbuildAPI.Build, error) {
	builds, err := buildClient.Builds(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for i := range builds.Items {
		var build = builds.Items[i]

		if build.Spec.Source == nil {
			continue
		}

		if build.Spec.Source.Type != shpbuildAPI.GitType {
			continue
		}

		if build.Spec.Source.Git == nil {
			continue
		}

		if strings.Contains(build.Spec.Source.Git.URL, repo) {
			return &build, nil
		}
	}

	return nil, nil
}

func findBuildRun(ctx context.Context, buildClient *shpbuild.ShipwrightV1beta1Client, namespace string, repo string) (*shpbuildAPI.BuildRun, error) {
	buildRuns, err := buildClient.BuildRuns(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var candidate *shpbuildAPI.BuildRun

	for i := range buildRuns.Items {
		var buildRun = buildRuns.Items[i]

		if buildRun.Status.CompletionTime == nil {
			continue
		}

		if buildRun.Spec.Build.Spec == nil {
			continue
		}

		if buildRun.Spec.Build.Spec.Source == nil {
			continue
		}

		if buildRun.Spec.Build.Spec.Source.Type != shpbuildAPI.GitType {
			continue
		}

		if buildRun.Spec.Build.Spec.Source.Git == nil {
			continue
		}

		if strings.Contains(buildRun.Spec.Build.Spec.Source.Git.URL, repo) {
			switch {
			case candidate == nil:
				candidate = &buildRun

			case candidate.Status.CompletionTime.Before(buildRun.Status.CompletionTime):
				candidate = &buildRun
			}
		}
	}

	return candidate, nil
}

func findService(ctx context.Context, servingClient *serving.ServingV1Client, namespace string, image string) (*servingv1.Service, error) {
	services, err := servingClient.Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for i := range services.Items {
		var service = services.Items[i]

		if userImage, ok := service.Annotations[userImageAnnotation]; ok {
			if userImage == image {
				return &service, nil
			}
		}
	}

	return nil, nil
}

func waitForBuildRunCompletion(ctx context.Context, buildClient *shpbuild.ShipwrightV1beta1Client, buildRun *shpbuildAPI.BuildRun) (*shpbuildAPI.BuildRun, error) {
	var conditionFunc = func(ctx context.Context) (done bool, err error) {
		buildRun, err = buildClient.BuildRuns(buildRun.Namespace).Get(ctx, buildRun.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		var condition = buildRun.Status.GetCondition(shpbuildAPI.Succeeded)
		if condition == nil {
			return false, nil
		}

		switch condition.Status {
		case corev1.ConditionTrue:
			if buildRun.Status.CompletionTime != nil {
				return true, nil
			}

		case corev1.ConditionFalse:
			return false, fmt.Errorf(condition.Message)
		}

		return false, nil
	}

	timeout := lookUpTimeout(ctx, buildClient, buildRun)
	deadlineCtx, deadlineCancel := context.WithTimeout(ctx, timeout)
	defer deadlineCancel()

	if err := wait.PollUntilContextTimeout(deadlineCtx, 10*time.Second, timeout, true, conditionFunc); err != nil {
		return nil, fmt.Errorf("waiting for build-run to finish failed: %w", err)
	}

	return buildRun, nil
}

func lookUpTimeout(ctx context.Context, buildClient *shpbuild.ShipwrightV1beta1Client, buildRun *shpbuildAPI.BuildRun) time.Duration {
	if buildRun.Spec.Timeout != nil {
		return buildRun.Spec.Timeout.Duration
	}

	if buildRun.Spec.Build.Name != nil {
		build, err := buildClient.Builds(buildRun.Namespace).Get(ctx, *buildRun.Spec.Build.Name, metav1.GetOptions{})
		if err == nil {
			if build.Spec.Timeout != nil {
				return build.Spec.Timeout.Duration
			}
		}
	}

	if buildRun.Spec.Build.Spec != nil {
		if buildRun.Spec.Build.Spec.Timeout != nil {
			return buildRun.Spec.Build.Spec.Timeout.Duration
		}
	}

	return defaultBuildRunWaitTimeout
}

func nudgeService(ctx context.Context, servingClient *serving.ServingV1Client, service *servingv1.Service) {
	log.Printf("nudging Knative servive %s to force a new revision\n", service.Name)

	// set update timestamp to force update
	annotations := service.Spec.Template.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[updateTimestampAnnotationKey] = time.Now().UTC().Format(time.RFC3339)
	service.Spec.Template.SetAnnotations(annotations)

	_, err := servingClient.Services(service.Namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("failed to update service %s to complete: %v\n", service.Name, err)
	}
}

func handleRepo(ctx context.Context, wg *sync.WaitGroup, repo string) int {
	namespace, buildClient, servingClient, err := inClusterSetup()
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}

	// --- --- ---

	build, err := findBuild(ctx, buildClient, namespace, repo)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}

	if build != nil {
		log.Printf("found build %s that references %s\n", build.Name, repo)
		return handleBuild(ctx, wg, servingClient, buildClient, build)
	}

	// --- --- ---

	buildRun, err := findBuildRun(ctx, buildClient, namespace, repo)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}

	if buildRun != nil {
		log.Printf("found standalone build-run %s that references %s\n", buildRun.Name, repo)
		return handleBuildRun(ctx, wg, servingClient, buildClient, buildRun)
	}

	// --- --- ---

	log.Printf("found no suitable build or standalone build-run in namespace %s for repository %s", namespace, repo)
	return http.StatusNotFound
}

func handleBuild(ctx context.Context, wg *sync.WaitGroup, servingClient *serving.ServingV1Client, buildClient *shpbuild.ShipwrightV1beta1Client, build *shpbuildAPI.Build) int {
	service, err := findService(ctx, servingClient, build.Namespace, build.Spec.Output.Image)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}

	if service == nil {
		log.Printf("found no service in namespace %s with image %s", build.Namespace, build.Spec.Output.Image)
		return http.StatusNotFound
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		newBuildRun := &shpbuildAPI.BuildRun{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("rebuild-%s-", build.Name),
				Namespace:    build.Namespace,
			},
			Spec: shpbuildAPI.BuildRunSpec{
				Build: shpbuildAPI.ReferencedBuild{
					Name: &build.Name,
				},
			},
		}

		log.Printf("creating build-run for build %s\n", build.Name)
		newBuildRun, err := buildClient.BuildRuns(build.Namespace).Create(ctx, newBuildRun, metav1.CreateOptions{})
		if err != nil {
			log.Printf("failed to create build-run for build %s: %v\n", build.Name, err)
			return
		}

		_, err = waitForBuildRunCompletion(ctx, buildClient, newBuildRun)
		if err != nil {
			log.Printf("failed to wait for build-run to complete: %v\n", err)
			return
		}

		nudgeService(ctx, servingClient, service)
	}()

	return http.StatusAccepted
}

func handleBuildRun(ctx context.Context, wg *sync.WaitGroup, servingClient *serving.ServingV1Client, buildClient *shpbuild.ShipwrightV1beta1Client, buildRun *shpbuildAPI.BuildRun) int {
	service, err := findService(ctx, servingClient, buildRun.Namespace, buildRun.Spec.Build.Spec.Output.Image)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}

	if service == nil {
		log.Printf("found no service in namespace %s with image %s", buildRun.Namespace, buildRun.Spec.Build.Spec.Output.Image)
		return http.StatusNotFound
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		rebuildBuildRun := &shpbuildAPI.BuildRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("rebuild-build-run-%s", time.Now().Format("20060102150405")),
				Namespace: buildRun.Namespace,
			},
			Spec: *buildRun.Spec.DeepCopy(),
		}

		log.Printf("creating standalone build-run %s\n", rebuildBuildRun.Name)
		buildRun, err := buildClient.BuildRuns(buildRun.Namespace).Create(ctx, rebuildBuildRun, metav1.CreateOptions{})
		if err != nil {
			log.Printf("failed to create new standalone build-run: %v\n", err)
			return
		}

		_, err = waitForBuildRunCompletion(ctx, buildClient, buildRun)
		if err != nil {
			log.Printf("failed to wait for build-run to complete: %v\n", err)
			return
		}

		nudgeService(ctx, servingClient, service)
	}()

	return http.StatusAccepted
}
