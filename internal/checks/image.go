package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*ImageCheck)(nil)

func Image(zone scw.Zone, name string) *ImageCheck {
	return &ImageCheck{
		zone:      zone,
		imageName: name,
	}
}

type ImageCheck struct {
	zone      scw.Zone
	imageName string

	rootVolumeType *string
	size           *scw.Size
}

func (c *ImageCheck) RootVolumeType(rootVolumeType string) *ImageCheck {
	c.rootVolumeType = &rootVolumeType

	return c
}

func (c *ImageCheck) SizeInGb(size uint64) *ImageCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *ImageCheck) Check(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListImages(&instance.ListImagesRequest{
		Name:    &c.imageName,
		Zone:    c.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	if len(resp.Images) == 0 {
		return fmt.Errorf("image %s not found, no images found", c.imageName)
	}

	if len(resp.Images) > 1 {
		return fmt.Errorf("multiple images found with name %s", c.imageName)
	}

	image := resp.Images[0]

	if image.Name != c.imageName {
		return fmt.Errorf("image name %s does not match expected %s", image.Name, c.imageName)
	}

	if c.rootVolumeType != nil && string(image.RootVolume.VolumeType) != *c.rootVolumeType {
		return fmt.Errorf("image root volume type %s does not match expected %s", image.RootVolume.VolumeType, *c.rootVolumeType)
	}

	if c.size != nil && image.RootVolume.Size != *c.size {
		return fmt.Errorf("image size %d does not match expected %d", image.RootVolume.Size, *c.size)
	}

	return nil
}
