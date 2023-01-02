package wellknown

import (
	"io"
	"net/http"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// NodeInfo returns a NodeInfo REST resource.
func (s *Service) NodeInfo() *NodeInfo {
	return &NodeInfo{
		service: s,
	}
}

func (s *Service) Webfinger() *Webfinger {
	return &Webfinger{
		service: s,
	}
}

func (w *Service) HostMeta(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/xrd+xml")
	io.WriteString(rw, `<?xml version="1.0" encoding="UTF-8"?>
		<XRD xmlns="http://docs.oasis-open.org/ns/xri/xrd-1.0">
		<Subject>`+r.Host+`</Subject>
		<Link rel="lrdd" template="https://`+r.Host+`/.well-known/webfinger?resource={uri}"/>
		</XRD>`)
}
