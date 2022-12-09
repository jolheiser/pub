package mastodon

import (
	"net/http"

	"github.com/davecheney/m/m"
)

type Notifications struct {
	service *Service
}

func (n *Notifications) Index(w http.ResponseWriter, r *http.Request) {
	user, err := n.service.authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var notifications []m.Notification
	if err := n.service.DB().Where("account_id = ?", user.ID).Preload("Status").Preload("Status.Account").Find(&notifications).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp []any
	for _, n := range notifications {
		resp = append(resp, map[string]any{
			"id":         toString(n.ID),
			"type":       n.Type,
			"created_at": n.CreatedAt.UTC().Format("2006-01-02T15:04:05.006Z"),
			"account":    serialize(n.Status.Account),
			"status":     serializeStatus(n.Status),
		})
	}
	toJSON(w, resp)
}