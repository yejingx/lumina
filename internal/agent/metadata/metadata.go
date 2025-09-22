package metadata

import (
	"encoding/json"
	"errors"
	"strconv"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"

	"lumina/internal/dao"
)

const (
	agentInfoKey     = "agent_info"
	lastFetchTimeKey = "last_fetch_time"
	jobKeyPrefix     = "job:"
)

type AgentInfo struct {
	Uuid              *string `json:"uuid,omitempty"`
	Token             *string `json:"token,omitempty"`
	RegisterTime      *string `json:"registerTime,omitempty"`
	S3AccessKeyID     *string `json:"s3AccessKeyID"`
	S3SecretAccessKey *string `json:"s3SecretAccessKey"`
}

func (info *AgentInfo) Update(new *AgentInfo) {
	if new.Uuid != nil {
		info.Uuid = new.Uuid
	}
	if new.Token != nil {
		info.Token = new.Token
	}
	if new.RegisterTime != nil {
		info.RegisterTime = new.RegisterTime
	}
	if new.S3AccessKeyID != nil {
		info.S3AccessKeyID = new.S3AccessKeyID
	}
	if new.S3SecretAccessKey != nil {
		info.S3SecretAccessKey = new.S3SecretAccessKey
	}
}

type MetadataDB struct {
	db     *badger.DB
	logger *logrus.Entry
}

func NewMetadataDB(dir string, logger *logrus.Entry) (*MetadataDB, error) {
	db, err := badger.Open(badger.DefaultOptions(dir).WithLoggingLevel(badger.ERROR))
	if err != nil {
		return nil, err
	}
	return &MetadataDB{
		db:     db,
		logger: logger,
	}, nil
}

func (m *MetadataDB) Close() error {
	return m.db.Close()
}

func (m *MetadataDB) Get(key []byte) ([]byte, error) {
	var val []byte
	err := m.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (m *MetadataDB) Set(key, val []byte) error {
	return m.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (m *MetadataDB) Delete(key []byte) error {
	return m.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (m *MetadataDB) List(prefix []byte) ([]*badger.Item, error) {
	var items []*badger.Item
	err := m.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			items = append(items, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (m *MetadataDB) GetAgentInfo() (*AgentInfo, error) {
	val, err := m.Get([]byte(agentInfoKey))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, err
	}
	info := &AgentInfo{}
	err = json.Unmarshal(val, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (m *MetadataDB) UpdateAgentInfo(new *AgentInfo) error {
	return m.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(agentInfoKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				newVal, err2 := json.Marshal(new)
				if err2 != nil {
					return err2
				}
				return txn.Set([]byte(agentInfoKey), newVal)
			}
			return err
		}
		return item.Value(func(val []byte) error {
			old := &AgentInfo{}
			json.Unmarshal(val, old)
			old.Update(new)
			newVal, err := json.Marshal(old)
			if err != nil {
				return err
			}
			return txn.Set([]byte(agentInfoKey), newVal)
		})
	})
}

func (m *MetadataDB) GetLastFetchTime() (int64, error) {
	t, err := m.Get([]byte(lastFetchTimeKey))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return strconv.ParseInt(string(t), 10, 64)
}

func (m *MetadataDB) SetLastFetchTime(t int64) error {
	return m.Set([]byte(lastFetchTimeKey), []byte(strconv.FormatInt(t, 10)))
}

func (m *MetadataDB) DeleteJob(id string) error {
	return m.Delete([]byte(jobKeyPrefix + id))
}

func (m *MetadataDB) GetJob(id string) (*dao.JobSpec, error) {
	val, err := m.Get([]byte(jobKeyPrefix + id))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, err
	}
	job := &dao.JobSpec{}
	err = json.Unmarshal(val, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (m *MetadataDB) SetJob(id string, job *dao.JobSpec) error {
	val, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return m.Set([]byte(jobKeyPrefix+id), val)
}

func (m *MetadataDB) GetJobs() ([]*dao.JobSpec, error) {
	prefix := []byte(jobKeyPrefix)
	jobs := make([]*dao.JobSpec, 0, 10)
	err := m.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			job := &dao.JobSpec{}
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, job)
			})
			if err != nil {
				m.logger.WithError(err).Errorf("unmarshal job %s", item.Key())
			} else {
				jobs = append(jobs, job)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
