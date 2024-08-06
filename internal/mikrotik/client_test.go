package mikrotik

// import (
// 	"testing"

// 	"github.com/caarlos0/env/v11"
// 	"github.com/stretchr/testify/assert"
// 	"sigs.k8s.io/external-dns/endpoint"
// )

// func TestCRUDRecord(t *testing.T) {
// 	// Fetch configuration from environment variables
// 	config := &Config{}
// 	err := env.Parse(config)
// 	if err != nil {
// 		t.Fatalf("failed to parse config from environment variables: %v", err)
// 	}

// 	// Attempt connection
// 	client, err := NewMikrotikClient(config)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, client)

// 	// Define the endpoint to create
// 	newEndpoint := &endpoint.Endpoint{
// 		DNSName:    "new.example.com",
// 		RecordType: "A",
// 		Targets:    endpoint.Targets{"9.10.11.12"},
// 		RecordTTL:  3600,
// 	}

// 	// Fetch all records
// 	records1, err := client.GetAll()
// 	assert.Nil(t, err)
// 	assert.NotEmpty(t, records1)

// 	// Call the Create function -> should work
// 	record1, err := client.Create(newEndpoint)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, record1)

// 	// Call the Create function again -> should fail, record already exists
// 	record2, err := client.Create(newEndpoint)
// 	assert.NotNil(t, err)
// 	assert.Nil(t, record2)

// 	// Fetch all records after creation
// 	records2, err := client.GetAll()
// 	assert.Nil(t, err)
// 	assert.NotEmpty(t, records2)

// 	// Ensure new records list is longer than the old one by 1
// 	// and that the new record is present
// 	assert.True(t, len(records2)-len(records1) == 1)

// 	var found bool
// 	for _, rec := range records2 {
// 		if rec.Name == newEndpoint.DNSName && rec.Address == newEndpoint.Targets[0] {
// 			found = true
// 			break
// 		}
// 	}
// 	assert.True(t, found)

// 	// Delete record -> should work
// 	err = client.Delete(newEndpoint)
// 	assert.Nil(t, err)

// 	// Delete record again -> should fail, not found
// 	err = client.Delete(newEndpoint)
// 	assert.NotNil(t, err)
// }
