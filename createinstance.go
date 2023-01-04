package main

import (
	"fmt"

	"github.com/davecheney/pub/internal/models"
	"github.com/davecheney/pub/internal/snowflake"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateInstanceCmd struct {
	Domain      string `required:"" help:"domain name of the instance to create"`
	Title       string `required:"" help:"title of the instance to create"`
	Description string `required:"" help:"description of the instance to create"`
	AdminEmail  string `required:"" help:"email address of the admin account to create"`
}

func (c *CreateInstanceCmd) Run(ctx *Context) error {
	db, err := gorm.Open(ctx.Dialector, &ctx.Config)
	if err != nil {
		return err
	}

	kp, err := generateRSAKeypair()
	if err != nil {
		return err
	}

	passwd, err := bcrypt.GenerateFromPassword(kp.privateKey, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return withTransaction(db, func(tx *gorm.DB) error {
		admin := models.Actor{
			ID:          snowflake.Now(),
			Type:        "Service",
			URI:         fmt.Sprintf("https://%s/u/%s", c.Domain, "admin"),
			Name:        "admin",
			Domain:      c.Domain,
			DisplayName: "admin",
			Locked:      false,
			Note:        "The admin account for " + c.Domain,
			Avatar:      "https://avatars.githubusercontent.com/u/1024?v=4",
			Header:      "https://avatars.githubusercontent.com/u/1024?v=4",
			PublicKey:   kp.publicKey,
		}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}

		instance := models.Instance{
			ID:               snowflake.Now(),
			Domain:           c.Domain,
			SourceURL:        "https://github.com/davecheney/pub",
			Title:            c.Title,
			ShortDescription: c.Description,
			Description:      c.Description,
			Thumbnail:        "https://avatars.githubusercontent.com/u/1024?v=4",
			Rules: []models.InstanceRule{{
				Text: "No loafing",
			}},
		}
		if err := tx.Create(&instance).Error; err != nil {
			return err
		}

		var adminRole models.AccountRole
		if err := tx.Where("name = ?", "admin").FirstOrCreate(&adminRole, models.AccountRole{
			Name:        "admin",
			Position:    1,
			Permissions: 0xFFFFFFFF,
			Highlighted: true,
		}).Error; err != nil {
			return err
		}

		adminAccount := models.Account{
			ID:                snowflake.Now(),
			InstanceID:        instance.ID,
			ActorID:           admin.ID,
			Email:             c.AdminEmail,
			EncryptedPassword: passwd,
			PrivateKey:        kp.privateKey,
			RoleID:            adminRole.ID,
		}
		if err := tx.Create(&adminAccount).Error; err != nil {
			return err
		}

		return tx.Model(&instance).Update("admin_id", adminAccount.ID).Error
	})
}

func withTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {
	tx := db.Begin()
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
