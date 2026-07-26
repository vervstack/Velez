package vcn_api_impl

import (
	"context"
	"testing"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// Test that ListPeers returns empty list without error.
func TestListPeersReturnsEmpty(t *testing.T) {
	impl := &Impl{
		pipeliner:  nil,
		vpnService: nil,
	}

	req := &pb.ListPeers_Request{
		NamespaceKey: "test-namespace",
	}

	resp, err := impl.ListPeers(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if len(resp.Peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(resp.Peers))
	}
}
