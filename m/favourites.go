package m

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Favourite struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	AccountID uint `gorm:"uniqueIndex:idx_account_status"`
	StatusID  uint `gorm:"uniqueIndex:idx_account_status"`
}

func (f *Favourite) AfterCreate(tx *gorm.DB) error {
	var status Status
	if err := tx.Preload("Account").First(&status, f.StatusID).Error; err != nil {
		return err
	}
	status.FavouritesCount++
	return tx.Save(&status).Error
}

func (f *Favourite) AfterDelete(tx *gorm.DB) error {
	var status Status
	if err := tx.Preload("Account").First(&status, f.StatusID).Error; err != nil {
		return err
	}
	status.FavouritesCount--
	return tx.Save(&status).Error
}

type Favourites struct {
	db *gorm.DB
}

func (f *Favourites) Create(w http.ResponseWriter, r *http.Request) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	var token Token
	if err := f.db.Preload("Account").Where("access_token = ?", accessToken).First(&token).Error; err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	id := mux.Vars(r)["id"]
	var status Status
	if err := f.db.Joins("Account").First(&status, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	favourite := Favourite{
		AccountID: token.AccountID,
		StatusID:  status.ID,
	}
	if err := f.db.Create(&favourite).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status.FavouritesCount++
	w.Header().Set("Content-Type", "application/json")
	json.MarshalFull(w, status.serialize())
}

func (f *Favourites) Destroy(w http.ResponseWriter, r *http.Request) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	var token Token
	if err := f.db.Preload("Account").Where("access_token = ?", accessToken).First(&token).Error; err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	id := mux.Vars(r)["id"]
	var status Status
	if err := f.db.Joins("Account").First(&status, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var favourite Favourite
	if err := f.db.Where("account_id = ? AND status_id = ?", token.AccountID, status.ID).First(&favourite).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := f.db.Delete(&favourite).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status.FavouritesCount--
	w.Header().Set("Content-Type", "application/json")
	json.MarshalFull(w, status.serialize())
}