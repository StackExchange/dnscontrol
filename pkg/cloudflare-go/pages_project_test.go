package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testPagesProjectResponse = `
    {
        "name": "Test Pages Project",
        "id": "5a321fc7-3162-7d36-adce-1213996a7",
        "created_on": "2021-01-01T00:00:00Z",
        "subdomain": "test.pages.dev",
        "domains": [
          "testdomain.com",
          "testdomain.org"
        ],
		"production_branch": "main",
        "source": {
          "type": "github",
          "config": {
            "owner": "cloudflare",
            "repo_name": "pages-test",
            "production_branch": "main",
            "pr_comments_enabled": true,
            "deployments_enabled": true,
			"preview_deployment_setting": "custom",
			"preview_branch_includes": [
				"release/*",
				"production",
				"main"
			],
			"preview_branch_excludes": [
				"dependabot/*",
				"dev",
				"*/ignore"
			]
          }
        },
        "build_config": {
          "build_caching": true,
          "build_command": "npm run build",
          "destination_dir": "build",
          "root_dir": "/",
          "web_analytics_tag": "0ee1d926cd60d2618a108d4232a75b73",
          "web_analytics_token": "c05bb382259183db3a0a822b64c11459"
        },
        "deployment_configs": {
          "preview": {
            "env_vars": {
              "BUILD_VERSION": {
                "value": "1.2"
              },
              "ENV": {
                "value": "preview"
              },
			  "API_KEY": {
				"value": "",
				"type": "secret_text"
			  }
            },
			"compatibility_date": "2022-08-15",
			"compatibility_flags": ["preview_flag"]
          },
          "production": {
            "env_vars": {
              "BUILD_VERSION": {
                "value": "1.2"
              },
              "ENV": {
                "value": "production"
              },
			  "API_KEY": {
				"value": "",
				"type": "secret_text"
			  }
            },
			"d1_databases": {
				"D1_BINDING": {
					"id": "a94509c6-0757-43f3-b053-474b0ab10935"
				}
			},
			"kv_namespaces": {
				"KV_BINDING": {
				  "namespace_id": "5eb63bbbe01eeed093cb22bb8f5acdc3"
				}
			},
		  	"durable_object_namespaces": {
				"DO_BINDING": {
			  		"namespace_id": "5eb63bbbe01eeed093cb22bb8f5acdc3"
				}
		  	},
			"r2_buckets": {
				"R2_BINDING": {
					"name": "some-bucket"
				}
			},
			"services": {
				"SERVICE_BINDING": {
					"service": "some-worker",
					"environment": "production"
				}
			},
			"compatibility_date": "2022-08-15",
			"compatibility_flags": ["production_flag"],
			"fail_open": false,
			"always_use_latest_compatibility_date": false,
			"usage_model": "bundled",
			"placement": {
				"mode": "smart"
			}
          }
        },
        "latest_deployment": {
          "id": "c35216d1-ebac-1a3a-d56c-ad74e54e5",
          "short_id": "c35216d1",
          "project_id": "5a321fc7-3162-7d36-adce-1213996a7",
          "project_name": "pages-test",
          "environment": "preview",
		  "production_branch": "main",
          "url": "https://c35216d1.pages-test.pages.dev",
          "created_on": "2021-03-09T00:55:03.923456Z",
          "modified_on": "2021-03-09T00:58:59.045655Z",
          "aliases": [
            "https://branchname.pages-test.pages.dev"
          ],
          "latest_stage": {
            "name": "deploy",
            "started_on": "2021-03-09T00:55:03.923456Z",
            "ended_on": "2021-03-09T00:58:59.045655Z",
            "status": "success"
          },
          "env_vars": {
            "BUILD_VERSION": {
              "value": "1.2"
            },
            "ENV": {
              "value": "STAGING"
            },
			"API_KEY": {
			  "value": "",
			  "type": "secret_text"
			}
          },
		  "placement": {
			"mode": "smart"
		  },
		  "compatibility_date": "2022-08-15",
		  "compatibility_flags": ["deployment_flag"],
		  "fail_open": false,
		  "always_use_latest_compatibility_date": false,
		  "usage_model": "bundled",
          "deployment_trigger": {
            "type": "ad_hoc",
            "metadata": {
              "branch": "main",
              "commit_hash": "fa26be19de6bff93f70bc2308434e4a440bbad02",
              "commit_message": "Update index.html"
            }
          },
          "stages": [
            {
              "name": "queued",
              "started_on": "2021-06-03T15:38:15.608194Z",
              "ended_on": "2021-06-03T15:39:03.134378Z",
              "status": "active"
            },
            {
              "name": "test_stage_1",
              "started_on": null,
              "ended_on": null,
              "status": "idle"
            }
          ],
          "build_config": {
            "build_caching": true,
            "build_command": "npm run build",
            "destination_dir": "build",
            "root_dir": "/",
            "web_analytics_tag": "0ee1d926cd60d2618a108d4232a75b73",
            "web_analytics_token": "c05bb382259183db3a0a822b64c11459"
          },
          "source": {
            "type": "github",
            "config": {
              "owner": "cloudflare",
              "repo_name": "pages-test",
              "production_branch": "main",
              "pr_comments_enabled": true,
              "deployments_enabled": true,
			  "preview_deployment_setting": "custom",
			  "preview_branch_includes": [
				"release/*",
				"production",
				"main"
			  ],
			  "preview_branch_excludes": [
				"dependabot/*",
				"dev",
				"*/ignore"
			  ]
            }
          }
        },
        "canonical_deployment": {
          "id": "c35216d1-ebac-1a3a-d56c-ad74e54e5",
          "short_id": "c35216d1",
          "project_id": "5a321fc7-3162-7d36-adce-1213996a7",
          "project_name": "pages-test",
          "environment": "preview",
          "url": "https://c35216d1.pages-test.pages.dev",
          "created_on": "2021-03-09T00:55:03.923456Z",
          "modified_on": "2021-03-09T00:58:59.045655Z",
		  "production_branch": "main",
          "aliases": [
            "https://branchname.pages-test.pages.dev"
          ],
          "latest_stage": {
            "name": "deploy",
            "started_on": "2021-03-09T00:55:03.923456Z",
            "ended_on": "2021-03-09T00:58:59.045655Z",
            "status": "success"
          },
          "env_vars": {
            "BUILD_VERSION": {
              "value": "1.2"
            },
            "ENV": {
              "value": "STAGING"
            },
			"API_KEY": {
			  "value": "",
			  "type": "secret_text"
			}
          },
		  "placement": {
			"mode": "smart"
		  },
		  "compatibility_date": "2022-08-15",
		  "compatibility_flags": ["deployment_flag"],
		  "fail_open": false,
		  "always_use_latest_compatibility_date": false,
		  "build_image_major_version": 1,
		  "usage_model": "bundled",
          "deployment_trigger": {
            "type": "ad_hoc",
            "metadata": {
              "branch": "main",
              "commit_hash": "fa26be19de6bff93f70bc2308434e4a440bbad02",
              "commit_message": "Update index.html"
            }
          },
          "stages": [
            {
              "name": "queued",
              "started_on": "2021-06-03T15:38:15.608194Z",
              "ended_on": "2021-06-03T15:39:03.134378Z",
              "status": "active"
            },
            {
              "name": "test_stage_1",
              "started_on": null,
              "ended_on": null,
              "status": "idle"
            }
          ],
          "build_config": {
            "build_caching": true,
            "build_command": "npm run build",
            "destination_dir": "build",
            "root_dir": "/",
            "web_analytics_tag": "0ee1d926cd60d2618a108d4232a75b73",
            "web_analytics_token": "c05bb382259183db3a0a822b64c11459"
          },
          "source": {
            "type": "github",
            "config": {
              "owner": "cloudflare",
              "repo_name": "pages-test",
              "production_branch": "main",
              "pr_comments_enabled": true,
              "deployments_enabled": true,
			  "preview_deployment_setting": "custom",
			  "preview_branch_includes": [
				"release/*",
				"production",
				"main"
			  ],
			  "preview_branch_excludes": [
				"dependabot/*",
				"dev",
				"*/ignore"
			  ]
            }
          }
        }
      }`
)

var (
	pagesProjectCreatedOn, _ = time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")

	expectedPagesProject = &PagesProject{
		SubDomain: "test.pages.dev",
		Name:      "Test Pages Project",
		Domains: []string{
			"testdomain.com",
			"testdomain.org",
		},
		CanonicalDeployment: *expectedPagesProjectDeployment,
		BuildConfig:         *expectedPagesProjectBuildConfig,
		CreatedOn:           &pagesProjectCreatedOn,
		DeploymentConfigs:   *expectedPagesProjectDeploymentConfigs,
		Source:              expectedPagesProjectSource,
		ID:                  "5a321fc7-3162-7d36-adce-1213996a7",
		LatestDeployment:    *expectedPagesProjectDeployment,
		ProductionBranch:    "main",
	}

	deploymentCreatedOn, _  = time.Parse(time.RFC3339, "2021-03-09T00:55:03.923456Z")
	deploymentModifiedOn, _ = time.Parse(time.RFC3339, "2021-03-09T00:58:59.045655Z")

	expectedPagesProjectDeployment = &PagesProjectDeployment{
		ID:          "c35216d1-ebac-1a3a-d56c-ad74e54e5",
		ShortID:     "c35216d1",
		ProjectID:   "5a321fc7-3162-7d36-adce-1213996a7",
		ProjectName: "pages-test",
		Environment: "preview",
		URL:         "https://c35216d1.pages-test.pages.dev",
		CreatedOn:   &deploymentCreatedOn,
		ModifiedOn:  &deploymentModifiedOn,
		Aliases: []string{
			"https://branchname.pages-test.pages.dev",
		},
		LatestStage: *expectedPagesProjectLatestDeploymentStage,
		EnvVars: EnvironmentVariableMap{
			"BUILD_VERSION": &EnvironmentVariable{
				Value: "1.2",
			},
			"ENV": &EnvironmentVariable{
				Value: "STAGING",
			},
			"API_KEY": &EnvironmentVariable{
				Value: "",
				Type:  SecretText,
			},
		},
		Placement: &Placement{
			Mode: PlacementModeSmart,
		},
		CompatibilityFlags: []string{"deployment_flag"},
		CompatibilityDate:  "2022-08-15",
		UsageModel:         Bundled,
		DeploymentTrigger:  *expectedPagesProjectDeploymentTrigger,
		Stages:             expectedStages,
		BuildConfig:        *expectedPagesProjectBuildConfig,
		Source:             *expectedPagesProjectSource,
		ProductionBranch:   "main",
	}

	latestDeploymentStageStartedOn, _ = time.Parse(time.RFC3339, "2021-03-09T00:55:03.923456Z")
	latestDeploymentStageEndedOn, _   = time.Parse(time.RFC3339, "2021-03-09T00:58:59.045655Z")

	expectedPagesProjectLatestDeploymentStage = &PagesProjectDeploymentStage{
		Name:      "deploy",
		StartedOn: &latestDeploymentStageStartedOn,
		EndedOn:   &latestDeploymentStageEndedOn,
		Status:    "success",
	}

	expectedPagesProjectDeploymentTrigger = &PagesProjectDeploymentTrigger{
		Type:     "ad_hoc",
		Metadata: expectedPagesProjectDeploymentTriggerMetadata,
	}

	expectedPagesProjectDeploymentTriggerMetadata = &PagesProjectDeploymentTriggerMetadata{
		Branch:        "main",
		CommitHash:    "fa26be19de6bff93f70bc2308434e4a440bbad02",
		CommitMessage: "Update index.html",
	}

	queuedStageStartedOn, _ = time.Parse(time.RFC3339, "2021-06-03T15:38:15.608194Z")
	queuedStageEndedOn, _   = time.Parse(time.RFC3339, "2021-06-03T15:39:03.134378Z")

	expectedStages = []PagesProjectDeploymentStage{
		{
			Name:      "queued",
			StartedOn: &queuedStageStartedOn,
			EndedOn:   &queuedStageEndedOn,
			Status:    "active",
		},
		{
			Name:   "test_stage_1",
			Status: "idle",
		},
	}

	expectedPagesProjectBuildConfig = &PagesProjectBuildConfig{
		BuildCaching:      BoolPtr(true),
		BuildCommand:      "npm run build",
		DestinationDir:    "build",
		RootDir:           "/",
		WebAnalyticsTag:   "0ee1d926cd60d2618a108d4232a75b73",
		WebAnalyticsToken: "c05bb382259183db3a0a822b64c11459",
	}

	expectedPagesProjectDeploymentConfigs = &PagesProjectDeploymentConfigs{
		Preview:    *expectedPagesProjectDeploymentConfigPreview,
		Production: *expectedPagesProjectDeploymentConfigProduction,
	}

	expectedPagesProjectDeploymentConfigPreview = &PagesProjectDeploymentConfigEnvironment{
		EnvVars: EnvironmentVariableMap{
			"BUILD_VERSION": &EnvironmentVariable{
				Value: "1.2",
			},
			"ENV": &EnvironmentVariable{
				Value: "preview",
			},
			"API_KEY": &EnvironmentVariable{
				Value: "",
				Type:  SecretText,
			},
		},
		CompatibilityDate:  "2022-08-15",
		CompatibilityFlags: []string{"preview_flag"},
	}

	expectedPagesProjectDeploymentConfigProduction = &PagesProjectDeploymentConfigEnvironment{
		EnvVars: EnvironmentVariableMap{
			"BUILD_VERSION": &EnvironmentVariable{
				Value: "1.2",
			},
			"ENV": &EnvironmentVariable{
				Value: "production",
			},
			"API_KEY": &EnvironmentVariable{
				Value: "",
				Type:  SecretText,
			},
		},
		KvNamespaces: NamespaceBindingMap{
			"KV_BINDING": &NamespaceBindingValue{Value: "5eb63bbbe01eeed093cb22bb8f5acdc3"},
		},
		D1Databases: D1BindingMap{
			"D1_BINDING": &D1Binding{ID: "a94509c6-0757-43f3-b053-474b0ab10935"},
		},
		DoNamespaces: NamespaceBindingMap{
			"DO_BINDING": &NamespaceBindingValue{Value: "5eb63bbbe01eeed093cb22bb8f5acdc3"},
		},
		R2Bindings: R2BindingMap{
			"R2_BINDING": &R2BindingValue{Name: "some-bucket"},
		},
		ServiceBindings: ServiceBindingMap{
			"SERVICE_BINDING": &ServiceBinding{
				Service:     "some-worker",
				Environment: "production",
			},
		},
		CompatibilityDate:                "2022-08-15",
		CompatibilityFlags:               []string{"production_flag"},
		FailOpen:                         false,
		AlwaysUseLatestCompatibilityDate: false,
		UsageModel:                       Bundled,
		Placement: &Placement{
			Mode: PlacementModeSmart,
		},
	}

	expectedPagesProjectSource = &PagesProjectSource{
		Type:   "github",
		Config: expectedPagesProjectSourceConfig,
	}

	expectedPagesProjectSourceConfig = &PagesProjectSourceConfig{
		Owner:                    "cloudflare",
		RepoName:                 "pages-test",
		ProductionBranch:         "main",
		PRCommentsEnabled:        true,
		DeploymentsEnabled:       true,
		PreviewDeploymentSetting: PagesPreviewCustomBranches,
		PreviewBranchIncludes:    []string{"release/*", "production", "main"},
		PreviewBranchExcludes:    []string{"dependabot/*", "dev", "*/ignore"},
	}
)

func TestListPagesProjects(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": [
                %s
            ],
            "result_info": {
                "page": 1,
                "per_page": 100,
                "count": 1,
                "total_count": 1
            }
        }`, testPagesProjectResponse)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects", handler)

	expectedPagesProjects := []PagesProject{
		*expectedPagesProject,
	}
	expectedResultInfo := ResultInfo{
		Page:    1,
		PerPage: 100,
		Count:   1,
		Total:   1,
	}

	_, _, err := client.ListPagesProjects(context.Background(), AccountIdentifier(""), ListPagesProjectsParams{})
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}

	actual, resultInfo, err := client.ListPagesProjects(context.Background(), AccountIdentifier(testAccountID), ListPagesProjectsParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPagesProjects, actual)
		assert.Equal(t, expectedResultInfo, resultInfo)
	}
}

func TestPagesProject(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": %s
        }`, testPagesProjectResponse)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/Test Pages Project", handler)

	_, err := client.GetPagesProject(context.Background(), AccountIdentifier(""), "Test Pages Project")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}

	actual, err := client.GetPagesProject(context.Background(), AccountIdentifier(testAccountID), "Test Pages Project")
	if assert.NoError(t, err) {
		assert.Equal(t, *expectedPagesProject, actual)
	}
}

func TestCreatePagesProject(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": %s
        }`, testPagesProjectResponse)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects", handler)

	params := &CreatePagesProjectParams{
		SubDomain: "test.pages.dev",
		Name:      "Test Pages Project",
		Domains: []string{
			"testdomain.com",
			"testdomain.org",
		},
		CanonicalDeployment: *expectedPagesProjectDeployment,
		BuildConfig:         *expectedPagesProjectBuildConfig,
		DeploymentConfigs:   *expectedPagesProjectDeploymentConfigs,
		Source:              expectedPagesProjectSource,
		LatestDeployment:    *expectedPagesProjectDeployment,
		ProductionBranch:    "main",
	}
	_, err := client.CreatePagesProject(context.Background(), AccountIdentifier(""), *params)
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}

	actual, err := client.CreatePagesProject(context.Background(), AccountIdentifier(testAccountID), *params)
	if assert.NoError(t, err) {
		assert.Equal(t, *expectedPagesProject, actual)
	}
}

func TestUpdatePagesProject(t *testing.T) {
	setup()
	defer teardown()

	updateAttributes := &UpdatePagesProjectParams{
		ID:   "Test Pages Project",
		Name: "updated-project-name",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": %s
        }`, testPagesProjectResponse)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/Test Pages Project", handler)

	_, err := client.UpdatePagesProject(context.Background(), AccountIdentifier(""), *updateAttributes)
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}

	_, err = client.UpdatePagesProject(context.Background(), AccountIdentifier(testAccountID), *updateAttributes)

	assert.NoError(t, err)
}

func TestDeletePagesProject(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": null
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/Test Pages Project", handler)

	err := client.DeletePagesProject(context.Background(), AccountIdentifier(""), "Test Pages Project")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}

	err = client.DeletePagesProject(context.Background(), AccountIdentifier(testAccountID), "Test Pages Project")
	assert.NoError(t, err)
}
