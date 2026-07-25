package dropbox

import "context"

// FileMetadata is the subset of Dropbox file metadata this project cares about.
type FileMetadata struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PathLower string `json:"path_lower"`
}

// GetMetadata looks up a file by its stable Dropbox ID (or path).
// Returns (nil, nil) if the file does not exist (deleted or moved out of reach).
func (c *Client) GetMetadata(ctx context.Context, pathOrID string) (*FileMetadata, error) {
	var meta FileMetadata
	err := c.rpc(ctx, "/files/get_metadata", map[string]interface{}{
		"path": pathOrID,
	}, &meta)
	if err != nil {
		var dbxErr *Err
		if ok := asErr(err, &dbxErr); ok && dbxErr.StatusCode == 409 && dbxErr.HasTag("not_found") {
			return nil, nil
		}
		return nil, err
	}
	return &meta, nil
}

// ListFolder lists all files (non-recursively) directly under folderPath, following pagination.
func (c *Client) ListFolder(ctx context.Context, folderPath string) ([]FileMetadata, error) {
	var page struct {
		Entries []struct {
			Tag       string `json:".tag"`
			ID        string `json:"id"`
			Name      string `json:"name"`
			PathLower string `json:"path_lower"`
		} `json:"entries"`
		Cursor  string `json:"cursor"`
		HasMore bool   `json:"has_more"`
	}

	var files []FileMetadata

	err := c.rpc(ctx, "/files/list_folder", map[string]interface{}{
		"path": folderPath,
	}, &page)
	if err != nil {
		return nil, err
	}
	for _, e := range page.Entries {
		if e.Tag == "file" {
			files = append(files, FileMetadata{ID: e.ID, Name: e.Name, PathLower: e.PathLower})
		}
	}

	for page.HasMore {
		var cont struct {
			Entries []struct {
				Tag       string `json:".tag"`
				ID        string `json:"id"`
				Name      string `json:"name"`
				PathLower string `json:"path_lower"`
			} `json:"entries"`
			Cursor  string `json:"cursor"`
			HasMore bool   `json:"has_more"`
		}
		err := c.rpc(ctx, "/files/list_folder/continue", map[string]interface{}{
			"cursor": page.Cursor,
		}, &cont)
		if err != nil {
			return nil, err
		}
		for _, e := range cont.Entries {
			if e.Tag == "file" {
				files = append(files, FileMetadata{ID: e.ID, Name: e.Name, PathLower: e.PathLower})
			}
		}
		page.HasMore = cont.HasMore
		page.Cursor = cont.Cursor
	}

	return files, nil
}

// GetOrCreateSharedLink returns the current shareable URL for a file, creating
// one if it doesn't already have one.
func (c *Client) GetOrCreateSharedLink(ctx context.Context, pathOrID string) (string, error) {
	var listResp struct {
		Links []struct {
			URL string `json:"url"`
		} `json:"links"`
	}
	if err := c.rpc(ctx, "/sharing/list_shared_links", map[string]interface{}{
		"path":        pathOrID,
		"direct_only": true,
	}, &listResp); err != nil {
		return "", err
	}
	if len(listResp.Links) > 0 {
		return listResp.Links[0].URL, nil
	}

	var createResp struct {
		URL string `json:"url"`
	}
	err := c.rpc(ctx, "/sharing/create_shared_link_with_settings", map[string]interface{}{
		"path": pathOrID,
	}, &createResp)
	if err == nil {
		return createResp.URL, nil
	}

	var dbxErr *Err
	if ok := asErr(err, &dbxErr); ok && dbxErr.StatusCode == 409 && dbxErr.HasTag("shared_link_already_exists") {
		// Someone/something created a link between our list and create calls; fetch it again.
		if err2 := c.rpc(ctx, "/sharing/list_shared_links", map[string]interface{}{
			"path":        pathOrID,
			"direct_only": true,
		}, &listResp); err2 == nil && len(listResp.Links) > 0 {
			return listResp.Links[0].URL, nil
		}
	}
	return "", err
}

// ResolveSharedLinkFileID resolves a dropbox.com share URL to the stable file ID
// of the underlying file, so it can be re-linked later even if renamed or moved.
func (c *Client) ResolveSharedLinkFileID(ctx context.Context, sharedURL string) (fileID string, err error) {
	var meta FileMetadata
	err = c.rpc(ctx, "/sharing/get_shared_link_metadata", map[string]interface{}{
		"url": sharedURL,
	}, &meta)
	if err != nil {
		return "", err
	}
	return meta.ID, nil
}

func asErr(err error, target **Err) bool {
	e, ok := err.(*Err)
	if ok {
		*target = e
	}
	return ok
}
