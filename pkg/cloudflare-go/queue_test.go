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
	testQueueID           = "6b7efc370ea34ded8327fa20698dfe3a"
	testQueueName         = "example-queue"
	testQueueConsumerName = "example-consumer"
)

func testQueue() Queue {
	CreatedOn, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	ModifiedOn, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	return Queue{
		ID:                  testQueueID,
		Name:                testQueueName,
		CreatedOn:           &CreatedOn,
		ModifiedOn:          &ModifiedOn,
		ProducersTotalCount: 1,
		Producers: []QueueProducer{
			{
				Service:     "example-producer",
				Environment: "production",
			},
		},
		ConsumersTotalCount: 1,
		Consumers: []QueueConsumer{
			{
				Service:     "example-consumer",
				Environment: "production",
				Settings: QueueConsumerSettings{
					BatchSize:   10,
					MaxRetires:  3,
					MaxWaitTime: 5000,
				},
			},
		},
	}
}

func testQueueConsumer() QueueConsumer {
	CreatedOn, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	return QueueConsumer{
		Service:     "example-consumer",
		Environment: "production",
		Settings: QueueConsumerSettings{
			BatchSize:   10,
			MaxRetires:  3,
			MaxWaitTime: 5000,
		},
		QueueName: testQueueName,
		CreatedOn: &CreatedOn,
	}
}

func TestQueue_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": null,
  "messages": null,
  "result": [
    {
      "queue_id": "6b7efc370ea34ded8327fa20698dfe3a",
      "queue_name": "example-queue",
      "created_on": "2023-01-01T00:00:00Z",
      "modified_on": "2023-01-01T00:00:00Z",
      "producers_total_count": 1,
      "producers": [
        {
          "service": "example-producer",
          "environment": "production"
        }
      ],
      "consumers_total_count": 1,
      "consumers": [
        {
          "service": "example-consumer",
          "environment": "production",
          "settings": {
            "batch_size": 10,
            "max_retries": 3,
            "max_wait_time_ms": 5000
          }
        }
      ]
    }
  ],
  "result_info": {
    "page": 1,
    "per_page": 100,
    "count": 1,
    "total_count": 1,
    "total_pages": 1
  }
}`)
	})

	_, _, err := client.ListQueues(context.Background(), AccountIdentifier(""), ListQueuesParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	result, _, err := client.ListQueues(context.Background(), AccountIdentifier(testAccountID), ListQueuesParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(result))
		assert.Equal(t, testQueue(), result[0])
	}
}

func TestQueue_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": null,
		  "messages": null,
		  "result": {
			"queue_id": "6b7efc370ea34ded8327fa20698dfe3a",
			"queue_name": "example-queue",
			"created_on": "2023-01-01T00:00:00Z",
			"modified_on": "2023-01-01T00:00:00Z"
		}
	}`)
	})
	_, err := client.CreateQueue(context.Background(), AccountIdentifier(""), CreateQueueParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.CreateQueue(context.Background(), AccountIdentifier(testAccountID), CreateQueueParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}
	results, err := client.CreateQueue(context.Background(), AccountIdentifier(testAccountID), CreateQueueParams{Name: "example-queue"})
	if assert.NoError(t, err) {
		CreatedOn, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		ModifiedOn, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		createdQueue := Queue{
			ID:         testQueueID,
			Name:       testQueueName,
			CreatedOn:  &CreatedOn,
			ModifiedOn: &ModifiedOn,
		}

		assert.Equal(t, createdQueue, results)
	}
}

func TestQueue_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues/%s", testAccountID, testQueueName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": null
		}`)
	})
	err := client.DeleteQueue(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	err = client.DeleteQueue(context.Background(), AccountIdentifier(testAccountID), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}

	err = client.DeleteQueue(context.Background(), AccountIdentifier(testAccountID), testQueueName)
	assert.NoError(t, err)
}

func TestQueue_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues/%s", testAccountID, testQueueID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
		{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": {
			"queue_id": "6b7efc370ea34ded8327fa20698dfe3a",
			"queue_name": "example-queue",
			"created_on": "2023-01-01T00:00:00Z",
			"modified_on": "2023-01-01T00:00:00Z",
			"producers_total_count": 1,
			"producers": [
			  {
				"service": "example-producer",
				"environment": "production"
			  }
			],
			"consumers_total_count": 1,
			"consumers": [
			  {
				"service": "example-consumer",
				"environment": "production",
				"settings": {
				  "batch_size": 10,
				  "max_retries": 3,
				  "max_wait_time_ms": 5000
				}
			  }
			]
		  }
		}`)
	})

	_, err := client.GetQueue(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.GetQueue(context.Background(), AccountIdentifier(testAccountID), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}

	result, err := client.GetQueue(context.Background(), AccountIdentifier(testAccountID), testQueueID)
	if assert.NoError(t, err) {
		assert.Equal(t, testQueue(), result)
	}
}

func TestQueue_Update(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues/%s", testAccountID, testQueueName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": null,
		  "messages": null,
		  "result": {
			"queue_id": "6b7efc370ea34ded8327fa20698dfe3a",
			"queue_name": "renamed-example-queue",
			"created_on": "2023-01-01T00:00:00Z",
			"modified_on": "2023-01-01T00:00:00Z"
		}
	}`)
	})
	_, err := client.UpdateQueue(context.Background(), AccountIdentifier(""), UpdateQueueParams{Name: testQueueName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.UpdateQueue(context.Background(), AccountIdentifier(testAccountID), UpdateQueueParams{Name: testQueueName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}

	results, err := client.UpdateQueue(context.Background(), AccountIdentifier(testAccountID), UpdateQueueParams{Name: testQueueName, UpdatedName: "renamed-example-queue"})
	if assert.NoError(t, err) {
		CreatedOn, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		ModifiedOn, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		createdQueue := Queue{
			ID:         testQueueID,
			Name:       "renamed-example-queue",
			CreatedOn:  &CreatedOn,
			ModifiedOn: &ModifiedOn,
		}

		assert.Equal(t, createdQueue, results)
	}
}

func TestQueue_ListConsumers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues/%s/consumers", testAccountID, testQueueName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
	{
		  "success": true,
		  "errors": null,
		  "messages": null,
		  "result": [
			{
			  "service": "example-consumer",
			  "environment": "production",
			  "settings": {
				"batch_size": 10,
				"max_retries": 3,
				"max_wait_time_ms": 5000
			  },
			  "queue_name": "example-queue",
			  "created_on": "2023-01-01T00:00:00Z"
			}
		  ],
		  "result_info": {
			"page": 1,
			"per_page": 100,
			"count": 1,
			"total_count": 1,
			"total_pages": 1
		  }
		}`)
	})

	_, _, err := client.ListQueueConsumers(context.Background(), AccountIdentifier(""), ListQueueConsumersParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, _, err = client.ListQueueConsumers(context.Background(), AccountIdentifier(testAccountID), ListQueueConsumersParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}

	result, _, err := client.ListQueueConsumers(context.Background(), AccountIdentifier(testAccountID), ListQueueConsumersParams{QueueName: testQueueName})
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(result))
		assert.Equal(t, testQueueConsumer(), result[0])
	}
}

func TestQueue_CreateConsumer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues/%s/consumers", testAccountID, testQueueName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": {
			"service": "example-consumer",
			"environment": "production",
			"settings": {
			  "batch_size": 10,
			  "max_retries": 3,
			  "max_wait_time_ms": 5000
			},
			"dead_letter_queue": "example-dlq",
			"queue_name": "example-queue",
			"created_on": "2023-01-01T00:00:00Z"
		  }
		}`)
	})

	_, err := client.CreateQueueConsumer(context.Background(), AccountIdentifier(""), CreateQueueConsumerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.CreateQueueConsumer(context.Background(), AccountIdentifier(testAccountID), CreateQueueConsumerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}

	result, err := client.CreateQueueConsumer(context.Background(), AccountIdentifier(testAccountID), CreateQueueConsumerParams{QueueName: testQueueName, Consumer: QueueConsumer{
		Service:     "example-consumer",
		Environment: "production",
	}})
	if assert.NoError(t, err) {
		expectedQueueConsumer := testQueueConsumer()
		expectedQueueConsumer.DeadLetterQueue = "example-dlq"
		assert.Equal(t, expectedQueueConsumer, result)
	}
}

func TestQueue_DeleteConsumer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues/%s/consumers/%s", testAccountID, testQueueName, testQueueConsumerName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": null
		}`)
	})

	err := client.DeleteQueueConsumer(context.Background(), AccountIdentifier(""), DeleteQueueConsumerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	err = client.DeleteQueueConsumer(context.Background(), AccountIdentifier(testAccountID), DeleteQueueConsumerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}

	err = client.DeleteQueueConsumer(context.Background(), AccountIdentifier(testAccountID), DeleteQueueConsumerParams{QueueName: testQueueName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueConsumerName, err)
	}

	err = client.DeleteQueueConsumer(context.Background(), AccountIdentifier(testAccountID), DeleteQueueConsumerParams{QueueName: testQueueName, ConsumerName: testQueueConsumerName})
	assert.NoError(t, err)
}

func TestQueue_UpdateConsumer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/queues/%s/consumers/%s", testAccountID, testQueueName, testQueueConsumerName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": {
			"service": "example-consumer",
			"environment": "production",
			"settings": {
			  "batch_size": 10,
			  "max_retries": 3,
			  "max_wait_time_ms": 5000
			},
			"queue_name": "example-queue",
			"created_on": "2023-01-01T00:00:00Z"
		  }
		}`)
	})

	_, err := client.UpdateQueueConsumer(context.Background(), AccountIdentifier(""), UpdateQueueConsumerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.UpdateQueueConsumer(context.Background(), AccountIdentifier(testAccountID), UpdateQueueConsumerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueName, err)
	}

	_, err = client.UpdateQueueConsumer(context.Background(), AccountIdentifier(testAccountID), UpdateQueueConsumerParams{QueueName: testQueueName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingQueueConsumerName, err)
	}

	result, err := client.UpdateQueueConsumer(context.Background(), AccountIdentifier(testAccountID), UpdateQueueConsumerParams{QueueName: testQueueName, Consumer: QueueConsumer{
		Name:        testQueueConsumerName,
		Service:     "example-consumer",
		Environment: "production",
	}})
	if assert.NoError(t, err) {
		assert.Equal(t, testQueueConsumer(), result)
	}
}
