package scm

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type repoLeaseRecord struct {
	gorm.Model
	CacheKey string `gorm:"column:cache_key;uniqueIndex;size:255;not null"`
	OwnerID  string `gorm:"column:owner_id;not null"`
	Token    string `gorm:"column:token;not null"`
}

func (repoLeaseRecord) TableName() string {
	return "scm_repo_leases"
}

type PostgresRepoLeaseManager struct {
	db *gorm.DB
}

func NewPostgresRepoLeaseManager(db *gorm.DB) (*PostgresRepoLeaseManager, error) {
	if db == nil {
		return nil, failures.WrapTerminal(errors.New("postgres repo lease manager db is required"))
	}
	if err := db.AutoMigrate(&repoLeaseRecord{}); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("migrate repo lease table: %w", err))
	}
	return &PostgresRepoLeaseManager{db: db}, nil
}

func (manager *PostgresRepoLeaseManager) Acquire(ctx context.Context, request applicationscm.RepoLeaseAcquireRequest) (domainscm.RepoLease, error) {
	if err := request.Validate(); err != nil {
		return domainscm.RepoLease{}, err
	}
	if manager == nil || manager.db == nil {
		return domainscm.RepoLease{}, failures.WrapTerminal(errors.New("postgres repo lease manager is not initialized"))
	}

	cacheKey := strings.TrimSpace(string(request.CacheKey))
	ownerID := strings.TrimSpace(request.OwnerID)
	token := strings.TrimSpace(request.Token)

	var existing repoLeaseRecord
	err := manager.db.WithContext(ctx).First(&existing, "cache_key = ?", cacheKey).Error
	if err == nil {
		if existing.OwnerID == ownerID && existing.Token == token {
			return domainscm.RepoLease{CacheKey: request.CacheKey, OwnerID: ownerID, Token: token}, nil
		}
		return domainscm.RepoLease{}, failures.WrapTransient(fmt.Errorf("repository cache lease already held for %s", request.CacheKey))
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domainscm.RepoLease{}, failures.WrapTransient(fmt.Errorf("load repo lease: %w", err))
	}

	record := repoLeaseRecord{CacheKey: cacheKey, OwnerID: ownerID, Token: token}
	if createErr := manager.db.WithContext(ctx).Create(&record).Error; createErr != nil {
		return domainscm.RepoLease{}, failures.WrapTransient(fmt.Errorf("create repo lease: %w", createErr))
	}
	return domainscm.RepoLease{CacheKey: request.CacheKey, OwnerID: ownerID, Token: token}, nil
}

func (manager *PostgresRepoLeaseManager) Release(ctx context.Context, lease domainscm.RepoLease) error {
	if err := lease.Validate(); err != nil {
		return err
	}
	if manager == nil || manager.db == nil {
		return failures.WrapTerminal(errors.New("postgres repo lease manager is not initialized"))
	}
	cacheKey := strings.TrimSpace(string(lease.CacheKey))
	ownerID := strings.TrimSpace(lease.OwnerID)
	token := strings.TrimSpace(lease.Token)
	if err := manager.db.WithContext(ctx).Unscoped().Where("cache_key = ? AND owner_id = ? AND token = ?", cacheKey, ownerID, token).Delete(&repoLeaseRecord{}).Error; err != nil {
		return failures.WrapTransient(fmt.Errorf("release repo lease: %w", err))
	}
	return nil
}

var _ applicationscm.RepoLeaseManager = (*PostgresRepoLeaseManager)(nil)
