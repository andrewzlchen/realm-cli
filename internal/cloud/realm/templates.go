package realm

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// Template represents an available Realm app template
type Template struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const (
	templatesPath             = adminAPI + "/templates"
	clientTemplatePathPattern = appPathPattern + "/templates/%s/client"
)

func (c *client) Templates() ([]Template, error) {
	res, resErr := c.do(
		http.MethodGet,
		templatesPath,
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get templates", res.StatusCode}
	}
	defer res.Body.Close()

	var templates []Template
	if err := json.NewDecoder(res.Body).Decode(&templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func (c *client) ClientTemplate(groupID, appID, templateID string) (*zip.Reader, error) {
	res, resErr := c.do(http.MethodGet, fmt.Sprintf(clientTemplatePathPattern, groupID, appID, templateID), api.RequestOptions{})
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get client template", res.StatusCode}
	}

	defer res.Body.Close()
	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	zipPkg, zipErr := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if zipErr != nil {
		return nil, zipErr
	}

	return zipPkg, nil
}
