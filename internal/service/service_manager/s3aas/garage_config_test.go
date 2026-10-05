package s3aas

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseGarageConfig_Values(t *testing.T) {
	content := []byte(`
metadata_dir = "/var/lib/garage/meta"
replication_factor = 3 # one copy per zone

[s3_api]
s3_region = "eu-west"
api_bind_addr = "[::]:3900"
`)

	region, replicationFactor := parseGarageConfig(content)

	require.Equal(t, "eu-west", region)
	require.EqualValues(t, 3, replicationFactor)
}

func Test_ParseGarageConfig_MissingKeys(t *testing.T) {
	region, replicationFactor := parseGarageConfig([]byte("# s3_region = \"commented\"\n"))

	require.Empty(t, region)
	require.Zero(t, replicationFactor)
}

const (
	testBucket   = "invoices"
	testOwnerKey = "billing:invoices"
)

func Test_ParseOwnerKeyName_Cases(t *testing.T) {
	cases := []struct {
		name       string
		keyName    string
		wantOwner  string
		wantBucket string
		wantIsKey  bool
	}{
		{"owner key", testOwnerKey, "billing", testBucket, true},
		{"colon in owner", "team:billing:invoices", "team:billing", testBucket, true},
		{"no separator", "plain-key", "", "", false},
		{"empty owner", ":invoices", "", "", false},
		{"empty bucket", "billing:", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner, bucket, isOwnerKey := parseOwnerKeyName(tc.keyName)

			require.Equal(t, tc.wantIsKey, isOwnerKey)
			require.Equal(t, tc.wantOwner, owner)
			require.Equal(t, tc.wantBucket, bucket)
		})
	}
}

func Test_OwnerFromKeyName_Cases(t *testing.T) {
	cases := []struct {
		name    string
		keyName string
		bucket  string
		want    string
	}{
		{"owner key", testOwnerKey, testBucket, "billing"},
		{"other bucket", testOwnerKey, "reports", ""},
		{"no separator", testBucket, testBucket, ""},
		{"empty owner", ":invoices", testBucket, ""},
		{"colon in owner", "team:billing:invoices", testBucket, "team:billing"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, ownerFromKeyName(tc.keyName, tc.bucket))
		})
	}
}
