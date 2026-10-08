package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

func Test_ParseS3OwnerKeyName_Cases(t *testing.T) {
	cases := []struct {
		name       string
		keyName    string
		wantOwner  string
		wantBucket string
		isOwnerKey bool
	}{
		{"owner and bucket", "svc:bucket", "svc", "bucket", true},
		{"splits on last separator", "a:b:c", "a:b", "c", true},
		{"no separator", "plain", "", "", false},
		{"empty owner", ":bucket", "", "", false},
		{"empty bucket", "svc:", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner, bucket, isOwnerKey := domain.ParseS3OwnerKeyName(tc.keyName)

			require.Equal(t, tc.wantOwner, owner)
			require.Equal(t, tc.wantBucket, bucket)
			require.Equal(t, tc.isOwnerKey, isOwnerKey)
		})
	}
}

func Test_S3OwnerKeyName_RoundTrips(t *testing.T) {
	owner, bucket, isOwnerKey := domain.ParseS3OwnerKeyName(domain.S3OwnerKeyName("svc", "photos"))

	require.True(t, isOwnerKey)
	require.Equal(t, "svc", owner)
	require.Equal(t, "photos", bucket)
}
