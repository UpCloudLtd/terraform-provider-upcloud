package database

import (
	"context"
	"testing"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
)

func TestIsManagedDatabaseFullyCreated(t *testing.T) {
	db := &upcloud.ManagedDatabase{
		Backups: make([]upcloud.ManagedDatabaseBackup, 0),
		State:   upcloud.ManagedDatabaseStatePoweroff,
		Users:   make([]upcloud.ManagedDatabaseUser, 0),
	}
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.State = upcloud.ManagedDatabaseStateRunning
	db.Backups = append(db.Backups, upcloud.ManagedDatabaseBackup{})
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.Users = append(db.Users, upcloud.ManagedDatabaseUser{})
	db.Backups = make([]upcloud.ManagedDatabaseBackup, 0)
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.Backups = append(db.Backups, upcloud.ManagedDatabaseBackup{})
	if !isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want true got false %+v", db)
	}
}

func TestWaitServiceNameToPropagate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err != nil {
		t.Errorf("waitServiceNameToPropagate failed %+v", err)
	}
}

func TestWaitServiceNameToPropagateContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err == nil {
		d, _ := ctx.Deadline()
		t.Errorf("waitServiceNameToPropagate failed didn't timeout before deadline %s", d.Format(time.RFC3339))
	}
}
