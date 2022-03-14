// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type PreresolvedImage struct {
	url     string
	origins []ctlconf.Origin
}

func NewPreresolvedImage(url string, origins []ctlconf.Origin) PreresolvedImage {
	return PreresolvedImage{url, copyAndAppendOrigins(origins)}
}

func (i PreresolvedImage) URL() (string, []ctlconf.Origin, error) {
	imageOrigins := copyAndAppendOrigins(i.origins, ctlconf.Origin{Preresolved: &ctlconf.OriginPreresolved{URL: i.url}})
	return i.url, imageOrigins, nil
}

func copyAndAppendOrigins(existing []ctlconf.Origin, new ...ctlconf.Origin) []ctlconf.Origin {
	all := make([]ctlconf.Origin, len(existing), len(existing)+len(new))
	copy(all, existing)
	return append(all, new...)
}
