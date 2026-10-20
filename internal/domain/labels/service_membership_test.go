package labels_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testService   = "app"
	testResource  = "app_db"
	testOtherName = "other"
)

func Test_ResourceTypeOf_Membership(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		labels      map[string]string
		container   string
		wantType    string
		wantBelongs bool
	}{
		{
			name:        "label matches and name has prefix",
			labels:      map[string]string{labels.VervServiceLabel: testService},
			container:   testResource,
			wantType:    "db",
			wantBelongs: true,
		},
		{
			name:        "label matches and name has no prefix",
			labels:      map[string]string{labels.VervServiceLabel: testService},
			container:   "web-1",
			wantType:    "",
			wantBelongs: true,
		},
		{
			name:        "label names another service despite matching prefix",
			labels:      map[string]string{labels.VervServiceLabel: testOtherName},
			container:   testResource,
			wantType:    "db",
			wantBelongs: false,
		},
		{
			name: "container labelled as its own service belongs to the service it is the web ui for",
			labels: map[string]string{
				labels.VervServiceLabel: testService + "_web_ui",
				labels.WebUiForLabel:    testService,
			},
			container:   testService + "_web_ui",
			wantType:    "web_ui",
			wantBelongs: true,
		},
		{
			name:        "unlabelled container falls back to name prefix",
			labels:      nil,
			container:   testResource,
			wantType:    "db",
			wantBelongs: true,
		},
		{
			name:        "unlabelled container without prefix does not belong",
			labels:      nil,
			container:   "other_db",
			wantType:    "",
			wantBelongs: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotType, gotBelongs := labels.ResourceTypeOf(testService, tc.labels, tc.container)

			require.Equal(t, tc.wantType, gotType)
			require.Equal(t, tc.wantBelongs, gotBelongs)
		})
	}
}

func Test_ServiceName_LabelFirstWithNameFallback(t *testing.T) {
	t.Parallel()

	withLabel := labels.ServiceName(map[string]string{labels.VervServiceLabel: testService}, []string{"/" + testResource})
	require.Equal(t, testService, withLabel)

	withoutLabel := labels.ServiceName(nil, []string{"/" + testResource})
	require.Equal(t, testResource, withoutLabel)

	require.Empty(t, labels.ServiceName(nil, nil))
}
