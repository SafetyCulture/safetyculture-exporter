// Copyright (c) 2020 SafetyCulture Pty Ltd. All Rights Reserved.

package update

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/hashicorp/go-version"
)

// ReleaseInfo is the details of an available release.
type ReleaseInfo struct {
	Version      string
	ChangelogURL string
}

// Check returns release info of a new version of this tool if available.
func Check(currentVersion string) *ReleaseInfo {
	ctx := context.Background()
	g := github.NewClient(&http.Client{})
	res, _, err := g.Repositories.GetLatestRelease(ctx, "SafetyCulture", "safetyculture-exporter")
	if err != nil {
		return nil
	}

	v := res.GetTagName()
	if VersionGreaterThanOrEqual(currentVersion, v) {
		return nil
	}

	return &ReleaseInfo{
		Version:      v,
		ChangelogURL: res.GetHTMLURL(),
	}
}

// VersionGreaterThanOrEqual returns true if version is greater or equal
func VersionGreaterThanOrEqual(v, w string) bool {
	vv, ve := version.NewVersion(v)
	vw, we := version.NewVersion(w)

	return ve == nil && we == nil && vv.GreaterThanOrEqual(vw)
}
