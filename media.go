package wikigo

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/saluja-ji/wikigo/models"
)

// GetFile gets metadata for a media file by its title (must start with "File:").
func (m *MediaClient) GetFile(ctx context.Context, title string) (*models.File, error) {
	escapedTitle := sanitizeTitle(title)
	req, err := m.client.newRequest(ctx, http.MethodGet, true, "/file/"+escapedTitle, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var file models.File
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, err
	}

	return &file, nil
}
