package m

import (
	"context"
	"time"

	"github.com/carlmjohnson/requests"
	"gorm.io/gorm"
)

type Instance struct {
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Domain           string `gorm:"uniqueIndex;size:64"`
	AdminID          *uint
	Admin            *Account
	SourceURL        string
	Title            string `gorm:"size:64"`
	ShortDescription string
	Description      string
	Thumbnail        string `gorm:"size:64"`
	AccountsCount    int    `gorm:"default:0;not null"`
	StatusesCount    int    `gorm:"default:0;not null"`

	DomainsCount int `gorm:"-"` // not stored

	Rules    []InstanceRule `gorm:"foreignKey:InstanceID"`
	Accounts []Account
}

func (i *Instance) AfterCreate(tx *gorm.DB) error {
	return i.updateAccountsCount(tx)
}

func (i *Instance) serializeNodeInfo() map[string]any {
	return map[string]any{
		"version": "2.0", // https://github.com/jhass/nodeinfo/blob/main/schemas/2.0/schema.json
		"software": map[string]any{
			"name":    "https://github.com/davecheney/m",
			"version": "0.0.0-devel",
		},
		"protocols": []string{
			"activitypub",
		},
		"services": map[string]any{
			"outbound": []any{},
			"inbound":  []any{},
		},
		"usage": map[string]any{
			"users": map[string]any{
				"total":          i.AccountsCount,
				"activeMonth":    0,
				"activeHalfyear": 0,
			},
			"localPosts": i.StatusesCount,
		},
		"openRegistrations": false,
		"metadata":          map[string]any{},
	}
}

func (i *Instance) updateAccountsCount(tx *gorm.DB) error {
	var count int64
	err := tx.Model(&Account{}).Where("instance_id = ?", i.ID).Count(&count).Error
	if err != nil {
		return err
	}
	return tx.Model(i).Update("accounts_count", count).Error
}

func (i *Instance) updateStatusesCount(tx *gorm.DB) error {
	var count int64
	err := tx.Model(&Status{}).Joins("Account").Where("instance_id = ?", i.ID).Count(&count).Error
	if err != nil {
		return err
	}
	return tx.Model(i).Update("statuses_count", count).Error
}

type InstanceRule struct {
	gorm.Model
	InstanceID uint
	Text       string
}

type instances struct {
	db *gorm.DB
}

func (i *instances) Count() (int, error) {
	var count int64
	if err := i.db.Model(&Instance{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (i *instances) FindByDomain(domain string) (*Instance, error) {
	var instance Instance
	if err := i.db.Model(&Instance{}).Preload("Admin").Preload("Admin.LocalAccount").Where("domain = ?", domain).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (i *instances) newRemoteInstanceFetcher() *RemoteInstanceFetcher {
	return &RemoteInstanceFetcher{}
}

type RemoteInstanceFetcher struct {
}

func (r *RemoteInstanceFetcher) Fetch(domain string) (*Instance, error) {
	obj, err := r.fetch(domain)
	if err != nil {
		return nil, err
	}
	return &Instance{
		Domain:           domain,
		Title:            stringFromAny(obj["title"]),
		ShortDescription: stringFromAny(obj["short_description"]),
		Description:      stringFromAny(obj["description"]),
	}, nil
}

func (r *RemoteInstanceFetcher) fetch(domain string) (map[string]any, error) {
	var obj map[string]any
	err := requests.URL("https://" + domain + "/api/v1/instance").ToJSON(&obj).Fetch(context.Background())
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (i *instances) FindOrCreate(domain string, fn func(string) (*Instance, error)) (*Instance, error) {
	var instance Instance
	err := i.db.Where("domain = ?", domain).First(&instance).Error
	if err == nil {
		return &instance, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	inst, err := fn(domain)
	if err != nil {
		return nil, err
	}
	if err := i.db.Create(inst).Error; err != nil {
		return nil, err
	}
	return inst, nil
}

func stringOrDefault(s string, def string) string {
	if s == "" {
		return def
	}
	return s
}
