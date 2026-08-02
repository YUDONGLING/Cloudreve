package inventory

import (
	"testing"

	"github.com/cloudreve/Cloudreve/v4/ent"
	entuser "github.com/cloudreve/Cloudreve/v4/ent/user"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
)

func TestIsValidShareChecksCurrentOwnerAccess(t *testing.T) {
	tests := []struct {
		name     string
		status   entuser.Status
		canShare bool
		wantErr  bool
	}{
		{name: "active owner with share permission", status: entuser.StatusActive, canShare: true},
		{name: "active owner without share permission", status: entuser.StatusActive, wantErr: true},
		{name: "manually banned owner", status: entuser.StatusManualBanned, canShare: true, wantErr: true},
		{name: "system banned owner", status: entuser.StatusSysBanned, canShare: true, wantErr: true},
		{name: "inactive owner", status: entuser.StatusInactive, canShare: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := &boolset.BooleanSet{}
			boolset.Set(types.GroupPermissionShare, tt.canShare, permissions)
			group := &ent.Group{Permissions: permissions}
			owner := &ent.User{ID: 1, Status: tt.status}
			owner.SetGroup(group)
			file := &ent.File{OwnerID: owner.ID, FileChildren: 1}
			share := &ent.Share{}
			share.SetUser(owner)
			share.SetFile(file)

			err := IsValidShare(share)
			if (err != nil) != tt.wantErr {
				t.Fatalf("IsValidShare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
