package mastodon

import (
	"net/http"

	"github.com/davecheney/pub/internal/httpx"
	"github.com/davecheney/pub/internal/models"
	"github.com/davecheney/pub/internal/to"
	"gorm.io/gorm"
)

func InstancesIndexV1(env *Env, w http.ResponseWriter, r *http.Request) error {
	return instancesIndex(env, w, r, serialiseInstanceV1)
}

func InstancesIndexV2(env *Env, w http.ResponseWriter, r *http.Request) error {
	return instancesIndex(env, w, r, serialiseInstanceV2)
}

func instancesIndex(env *Env, w http.ResponseWriter, r *http.Request, seraliser func(*models.Instance) map[string]any) error {
	var instance models.Instance
	if err := env.DB.Where("domain = ?", r.Host).Preload("Admin").Preload("Admin.Actor").Preload("Rules").Take(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return httpx.Error(http.StatusNotFound, err)
		}
		return err
	}
	if err := env.DB.Model(&models.Instance{}).Count(&instance.DomainsCount).Error; err != nil {
		return err
	}
	return to.JSON(w, seraliser(&instance))
}

func InstancesPeersShow(env *Env, w http.ResponseWriter, r *http.Request) error {
	var domains []string
	if err := env.DB.Model(&models.Actor{}).Group("Domain").Where("Domain != ?", r.Host).Pluck("domain", &domains).Error; err != nil {
		return err
	}
	return to.JSON(w, domains)
}

func InstancesActivityShow(env *Env, w http.ResponseWriter, r *http.Request) error {
	return to.JSON(w, []map[string]interface{}{})
}

func InstancesDomainBlocksShow(env *Env, w http.ResponseWriter, r *http.Request) error {
	return to.JSON(w, []map[string]interface{}{})
}
