package imagetags

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	hubPageSize = 100
	hubMaxPages = 10
)

type hubTagsPage struct {
	Next    string       `json:"next"`
	Results []hubTagInfo `json:"results"`
}

type hubTagInfo struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
	Images []struct {
		Digest string `json:"digest"`
	} `json:"images"`
}

func (t hubTagInfo) hasDigest(digest string) bool {
	if t.Digest == digest {
		return true
	}

	for _, img := range t.Images {
		if img.Digest == digest {
			return true
		}
	}

	return false
}

func (c *Client) hubTagsForDigest(ctx context.Context, repoPath, digest string) ([]string, error) {
	tags := make([]string, 0)
	pageUrl := fmt.Sprintf("%s/v2/repositories/%s/tags?page_size=%d", c.hubUrl, repoPath, hubPageSize)

	for i := 0; i < hubMaxPages && pageUrl != ""; i++ {
		page, err := c.fetchHubPage(ctx, pageUrl)
		if err != nil {
			return nil, rerrors.Wrap(err, "error fetching docker hub tags page")
		}

		for _, tag := range page.Results {
			if tag.hasDigest(digest) {
				tags = append(tags, tag.Name)
			}
		}

		pageUrl = page.Next
	}

	return tags, nil
}

func (c *Client) fetchHubPage(ctx context.Context, pageUrl string) (hubTagsPage, error) {
	var page hubTagsPage

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageUrl, nil)
	if err != nil {
		return page, rerrors.Wrap(err, "error building docker hub request")
	}

	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return page, rerrors.Wrap(err, "error requesting docker hub")
	}
	defer common.CloseWithLog(resp.Body.Close, "docker hub tags response body")

	if resp.StatusCode != http.StatusOK {
		return page, rerrors.Wrap(errHubUnexpectedStatus, fmt.Sprintf("unexpected status: %d", resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&page)
	if err != nil {
		return page, rerrors.Wrap(err, "error decoding docker hub tags response")
	}

	return page, nil
}
