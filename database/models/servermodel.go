package models

import (
	"time"

	"gorm.io/gorm"
)

type Server struct {
	gorm.Model
	Online      bool
	Host        string
	Port        string
	LastChecked time.Time `gorm:"autoCreateTime:nano"`
}

func GetServerCache(db *gorm.DB, host string, port string) Server {
	var server Server
	db.Where("host = ?", host).Where("port = ?", port).First(&server)
	if server.ID == 0 {
		server.Host = host
		server.Port = port
		server.Online = false
		server.LastChecked = time.Now().Add(-5 * time.Minute)
		db.Create(&server)
	}
	return server
}

func (s *Server) Update(db *gorm.DB) {
	var dbServ Server
	db.Where("id = ?", s.ID).First(&dbServ)

	dbServ.Online = s.Online
	dbServ.LastChecked = s.LastChecked
	db.Save(&dbServ)
}
