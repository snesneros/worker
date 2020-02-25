// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package kubernetes

import (
	"context"

	"github.com/go-vela/types/pipeline"
)

// CreateVolume creates the pipeline volume.
func (c *client) CreateVolume(ctx context.Context, b *pipeline.Build) error {
	return nil
}

// InspectVolume inspects the pipeline volume.
func (c *client) InspectVolume(ctx context.Context, b *pipeline.Build) ([]byte, error) {
	return nil, nil
}

// RemoveVolume deletes the pipeline volume.
func (c *client) RemoveVolume(ctx context.Context, b *pipeline.Build) error {
	return nil
}
