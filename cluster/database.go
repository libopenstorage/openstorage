package cluster

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

func readDatabase() (Database, error) {
	kvdb := kvdb.Instance()

	db := Database{Status: api.StatusInit,
		NodeEntries: make(map[string]NodeEntry)}

	kv, err := kvdb.Get("cluster/database")
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		logrus.Warn("Warning, could not read cluster database")
		return db, err
	}

	if kv == nil || bytes.Compare(kv.Value, []byte("{}")) == 0 {
		logrus.Info("Cluster is uninitialized...")
		return db, nil
	}
	if err := json.Unmarshal(kv.Value, &db); err != nil {
		logrus.Warn("Fatal, Could not parse cluster database ", kv)
		return db, err
	}

	return db, nil
}

func writeDatabase(db *Database) error {
	kvdb := kvdb.Instance()
	b, err := json.Marshal(db)
	if err != nil {
		logrus.Warnf("Fatal, Could not marshal cluster database to JSON: %v", err)
		return err
	}

	if _, err := kvdb.Put("cluster/database", b, 0); err != nil {
		logrus.Warnf("Fatal, Could not marshal cluster database to JSON: %v", err)
		return err
	}

	logrus.Info("Cluster database updated.")
	return nil
}
