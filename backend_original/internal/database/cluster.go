package database

import (
	"bome-backend/internal/config"
	"fmt"
	"log"
	"sync"
)

// DBCluster manages master and replica connections
type DBCluster struct {
	master   *DB
	replicas []*DB
	current  int
	mutex    sync.RWMutex
}

// NewDBCluster creates a new database cluster
func NewDBCluster(masterConfig, replica1Config, replica2Config *config.Config) (*DBCluster, error) {
	// Connect to master
	master, err := New(masterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %w", err)
	}

	// Connect to replicas
	replica1, err := New(replica1Config)
	if err != nil {
		log.Printf("Failed to connect to replica 1: %v", err)
		replica1 = nil
	}

	replica2, err := New(replica2Config)
	if err != nil {
		log.Printf("Failed to connect to replica 2: %v", err)
		replica2 = nil
	}

	replicas := []*DB{}
	if replica1 != nil {
		replicas = append(replicas, replica1)
	}
	if replica2 != nil {
		replicas = append(replicas, replica2)
	}

	return &DBCluster{
		master:   master,
		replicas: replicas,
		current:  0,
	}, nil
}

// GetMaster returns the master database for writes
func (cluster *DBCluster) GetMaster() *DB {
	return cluster.master
}

// GetReplica returns a replica database for reads (round-robin)
func (cluster *DBCluster) GetReplica() *DB {
	if len(cluster.replicas) == 0 {
		return cluster.master // Fallback to master
	}

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	replica := cluster.replicas[cluster.current]
	cluster.current = (cluster.current + 1) % len(cluster.replicas)
	return replica
}
