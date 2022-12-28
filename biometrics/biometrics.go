package biometrics

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"google.golang.org/api/option"
)

func Run(logger *log.Logger, _ string, db *sql.DB, mux *http.ServeMux, gcpClient *http.Client) error {
	logger = log.New(logger.Writer(), "[biometrics] ", logger.Flags())
	ctx := context.Background()

	visionClient, err := vision.NewImageAnnotatorClient(ctx, option.WithHTTPClient(gcpClient))
	if err != nil {
		return fmt.Errorf("can't build Cloud Vision API client: %w", err)
	}

	var bytes []byte
	resp, err := visionClient.AnnotateImage(ctx, &visionpb.AnnotateImageRequest{
		Image: &visionpb.Image{
			Content: bytes,
		},
	})
	if err != nil {
		return fmt.Errorf("can't annotate image: %w", err)
	}

	return nil
}
