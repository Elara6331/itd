package infinitime

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
)

type ResourceOperation int

const (
	// ResourceUpload represents the upload phase
	// of resource loading
	ResourceUpload = iota
	// ResourceRemove represents the obsolete
	// file removal phase of resource loading
	ResourceRemove
)

// resourceManifest is the structure of the resource manifest file
type resourceManifest struct {
	Resources []resource         `json:"resources"`
	Obsolete  []obsoleteResource `json:"obsolete_files"`
}

// resource represents a resource entry in the manifest
type resource struct {
	Name string `json:"filename"`
	Path string `json:"path"`
}

// obsoleteResource represents an obsolete file entry in the manifest
type obsoleteResource struct {
	Path  string `json:"path"`
	Since string `json:"since"`
}

// ResourceLoadProgress contains information on the progress of
// a resource load
type ResourceLoadProgress struct {
	Operation   ResourceOperation
	Name        string
	Total       uint32
	Transferred uint32
}

// LoadResources accepts the path of an InfiniTime resource archive and loads its contents to the watch's filesystem.
func LoadResources(archivePath string, fs *FS, progress func(ResourceLoadProgress)) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	manifestFl, err := r.Open("resources.json")
	if err != nil {
		return err
	}

	var manifest resourceManifest
	err = json.NewDecoder(manifestFl).Decode(&manifest)
	if err != nil {
		return err
	}

	err = manifestFl.Close()
	if err != nil {
		return err
	}

	for _, file := range manifest.Obsolete {
		err := fs.RemoveAll(file.Path)
		if err != nil {
			return err
		}

		progress(ResourceLoadProgress{
			Operation: ResourceRemove,
			Name:      filepath.Base(file.Path),
		})
	}

	for _, file := range manifest.Resources {
		src, err := r.Open(file.Name)
		if err != nil {
			return err
		}

		fi, err := src.Stat()
		if err != nil {
			return err
		}

		err = fs.MkdirAll(filepath.Dir(file.Path))
		if err != nil {
			return err
		}

		dst, err := fs.Create(file.Path, uint32(fi.Size()))
		if err != nil {
			return err
		}

		dst.ProgressFunc = func(transferred, total uint32) {
			progress(ResourceLoadProgress{
				Name:        file.Name,
				Transferred: transferred,
				Total:       total,
			})
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			return errors.Join(
				err,
				src.Close(),
				dst.Close(),
			)
		}

		err = src.Close()
		if err != nil {
			return err
		}

		err = dst.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
