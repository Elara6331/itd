package api

import (
	"github.com/mitchellh/mapstructure"
	"go.arsenm.dev/itd/internal/types"
)

func (c *Client) Rename(old, new string) error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeFS,
		Data: types.ReqDataFS{
			Type:  types.FSTypeMove,
			Files: []string{old, new},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Remove(paths ...string) error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeFS,
		Data: types.ReqDataFS{
			Type:  types.FSTypeDelete,
			Files: paths,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Mkdir(paths ...string) error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeFS,
		Data: types.ReqDataFS{
			Type:  types.FSTypeMkdir,
			Files: paths,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ReadDir(path string) ([]types.FileInfo, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeFS,
		Data: types.ReqDataFS{
			Type:  types.FSTypeList,
			Files: []string{path},
		},
	})
	if err != nil {
		return nil, err
	}
	var out []types.FileInfo
	err = mapstructure.Decode(res.Value, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ReadFile(path string) (string, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeFS,
		Data: types.ReqDataFS{
			Type:  types.FSTypeRead,
			Files: []string{path},
		},
	})
	if err != nil {
		return "", err
	}
	return res.Value.(string), nil
}

func (c *Client) WriteFile(path, data string) error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeFS,
		Data: types.ReqDataFS{
			Type:  types.FSTypeWrite,
			Files: []string{path},
			Data:  data,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
