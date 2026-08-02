package dbfs

import (
	"context"
	"testing"

	"github.com/cloudreve/Cloudreve/v4/ent"
	entuser "github.com/cloudreve/Cloudreve/v4/ent/user"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
)

func TestGetFileFromDirectLinkRejectsRestrictedOwner(t *testing.T) {
	tests := []struct {
		name      string
		status    entuser.Status
		batchSize int
	}{
		{name: "active owner without direct link permission", status: entuser.StatusActive},
		{name: "manually banned owner", status: entuser.StatusManualBanned, batchSize: 1},
		{name: "system banned owner", status: entuser.StatusSysBanned, batchSize: 1},
		{name: "inactive owner", status: entuser.StatusInactive, batchSize: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner := &ent.User{Status: tt.status}
			owner.SetGroup(&ent.Group{Settings: &types.GroupSetting{SourceBatchSize: tt.batchSize}})
			file := &ent.File{}
			file.SetOwner(owner)
			link := &ent.DirectLink{}
			link.SetFile(file)

			if _, err := (&DBFS{}).GetFileFromDirectLink(context.Background(), link); err == nil {
				t.Fatal("GetFileFromDirectLink() error = nil, want restricted owner to be rejected")
			}
		})
	}
}
