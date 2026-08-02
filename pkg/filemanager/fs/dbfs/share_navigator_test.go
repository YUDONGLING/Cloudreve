package dbfs

import (
	"context"
	"testing"

	"github.com/cloudreve/Cloudreve/v4/ent"
	entuser "github.com/cloudreve/Cloudreve/v4/ent/user"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/setting"
)

type restoredStateShareClient struct {
	inventory.ShareClient
	share *ent.Share
}

func (c *restoredStateShareClient) GetByHashID(context.Context, string) (*ent.Share, error) {
	return c.share, nil
}

func TestShareNavigatorRestoredStateRevalidatesRequester(t *testing.T) {
	ownerPermissions := &boolset.BooleanSet{}
	boolset.Set(types.GroupPermissionShare, true, ownerPermissions)
	owner := &ent.User{ID: 1, Status: entuser.StatusActive}
	owner.SetGroup(&ent.Group{Permissions: ownerPermissions})

	file := &ent.File{OwnerID: owner.ID, FileChildren: 1}
	share := &ent.Share{ID: 1, Props: &types.ShareProps{ShareView: true}}
	share.SetUser(owner)
	share.SetFile(file)

	hasher, err := hashid.New("restored-state-test")
	if err != nil {
		t.Fatalf("hashid.New() error = %v", err)
	}
	path, err := fs.NewUriFromString(fs.NewShareUri(hashid.EncodeShareID(hasher, share.ID), ""))
	if err != nil {
		t.Fatalf("fs.NewUriFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		canDownload bool
		wantErr     bool
	}{
		{name: "permission revoked", wantErr: true},
		{name: "permission retained", canDownload: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requesterPermissions := &boolset.BooleanSet{}
			boolset.Set(types.GroupPermissionShareDownload, tt.canDownload, requesterPermissions)
			requester := &ent.User{ID: 2, Status: entuser.StatusActive}
			requester.SetGroup(&ent.Group{Permissions: requesterPermissions})

			cachedOwner := &ent.User{ID: owner.ID, Status: entuser.StatusActive}
			root := newFile(nil, file)
			t.Cleanup(root.Recycle)
			root.OwnerModel = cachedOwner
			root.disableView = true
			root.CapabilitiesBs = &boolset.BooleanSet{0xff}

			navigator := NewShareNavigator(
				requester,
				nil,
				&restoredStateShareClient{share: share},
				nil,
				&setting.DBFS{},
				hasher,
			).(*shareNavigator)
			if err := navigator.RestoreState(shareNavigatorState{
				ShareRoot: root,
				OwnerRoot: root,
				Share:     share,
				Owner:     cachedOwner,
			}); err != nil {
				t.Fatalf("RestoreState() error = %v", err)
			}

			_, err := navigator.To(context.Background(), path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("To() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if root.OwnerModel != owner {
				t.Fatal("To() did not refresh the cached owner")
			}
			if root.disableView {
				t.Fatal("To() did not refresh the cached share view setting")
			}
			if root.CapabilitiesBs != shareNavigatorCapability {
				t.Fatal("To() did not refresh the cached navigator capabilities")
			}
		})
	}
}
