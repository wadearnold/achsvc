// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package pipeline

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/base"

	"github.com/moov-io/paygate/pkg/client"
	"github.com/moov-io/paygate/pkg/config"

	"github.com/moov-io/base/log"
)

// XferMerging represents logic for accepting ACH files to be merged together.
//
// The idea is to take Xfers and store them on a filesystem (or other durable storage)
// prior to a cutoff window. The specific storage could be based on the FileHeader.
//
// On the cutoff trigger WithEachMerged is called to merge files together and offer
// each merged file for an upload.
type XferMerging interface {
	HandleXfer(xfer Xfer) error
	HandleCancel(cancel CanceledTransfer) error

	WithEachMerged(func(*ach.File) error) (*processedTransfers, error)
}

func NewMerging(logger log.Logger, cfg config.Pipeline) (XferMerging, error) {
	dir := filepath.Join("storage", "mergable") // default directory
	if cfg.Merging != nil {
		dir = filepath.Join(cfg.Merging.Directory, "mergable")
	}

	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}

	return &filesystemMerging{
		baseDir: dir,
		cfg:     cfg.Merging,
		logger:  logger,
	}, nil
}

type filesystemMerging struct {
	baseDir string
	cfg     *config.Merging
	logger  log.Logger
}

func (m *filesystemMerging) HandleXfer(xfer Xfer) error {
	err1 := m.writeTransfer(xfer.Transfer)
	err2 := m.writeACHFile(xfer.Transfer.TransferID, xfer.File)

	if err1 != nil || err2 != nil {
		return fmt.Errorf("problem writing transfer: %v\n problem writing ACH file: %v", err1, err2)
	}

	return nil
}

func (m *filesystemMerging) writeTransfer(transfer *client.Transfer) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(transfer); err != nil {
		return err
	}

	path := filepath.Join(m.baseDir, fmt.Sprintf("%s.json", transfer.TransferID))
	if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (m *filesystemMerging) writeACHFile(transferID string, file *ach.File) error {
	var buf bytes.Buffer
	if err := ach.NewWriter(&buf).Write(file); err != nil {
		return err
	}

	path := filepath.Join(m.baseDir, fmt.Sprintf("%s.ach", transferID))
	if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (m *filesystemMerging) HandleCancel(cancel CanceledTransfer) error {
	path := filepath.Join(m.baseDir, fmt.Sprintf("%s.ach", cancel.TransferID))

	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		// file doesn't exist, so write one
		return ioutil.WriteFile(path+".canceled", nil, 0644)
	} else {
		// move the existing file
		return os.Rename(path, path+".canceled")
	}
}

func (m *filesystemMerging) isolateMergableDir() (string, error) {
	// rename m.baseDir so we're the only accessor for it, then recreate m.baseDir
	parent, _ := filepath.Split(m.baseDir)
	newdir := filepath.Join(parent, time.Now().Format("20060102-150405"))
	if err := os.Rename(m.baseDir, newdir); err != nil {
		return newdir, err
	}
	return newdir, os.Mkdir(m.baseDir, 0777) // create m.baseDir again
}

func getNonCanceledMatches(path string) ([]string, error) {
	positiveMatches, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}
	negativeMatches, err := filepath.Glob(path + "*.canceled")
	if err != nil {
		return nil, err
	}

	var out []string
	for i := range positiveMatches {
		exclude := false
		for j := range negativeMatches {
			// We match when a "XXX.ach.canceled" filepath exists and so we can't
			// include "XXX.ach" has a filepath from this function.
			if strings.HasPrefix(negativeMatches[j], positiveMatches[i]) {
				exclude = true
				break
			}
		}
		if !exclude {
			out = append(out, positiveMatches[i])
		}
	}
	return out, nil
}

type processedTransfers struct {
	transferIDs []string
}

func newProcessedTransfers(matches []string) *processedTransfers {
	processed := &processedTransfers{}

	for i := range matches {
		// each match follows $path/$transferID.ach so we can split that
		// and grab the transferID
		transferID := strings.TrimSuffix(filepath.Base(matches[i]), ".ach")
		processed.transferIDs = append(processed.transferIDs, transferID)
	}

	return processed
}

func (m *filesystemMerging) WithEachMerged(f func(*ach.File) error) (*processedTransfers, error) {
	// move the current directory so it's isolated and easier to debug later on
	dir, err := m.isolateMergableDir()
	if err != nil {
		return nil, fmt.Errorf("problem isolating newdir=%s error=%v", dir, err)
	}

	path := filepath.Join(dir, "*.ach")
	matches, err := getNonCanceledMatches(path)
	if err != nil {
		return nil, fmt.Errorf("problem with %s glob: %v", path, err)
	}

	var files []*ach.File
	var el base.ErrorList
	for i := range matches {
		file, err := ach.ReadFile(matches[i])
		if err != nil {
			el.Add(fmt.Errorf("problem reading %s: %v", matches[i], err))
			continue
		}
		if file != nil {
			files = append(files, file)
		}
	}
	files, err = ach.MergeFiles(files)
	if err != nil {
		el.Add(fmt.Errorf("unable to merge files: %v", err))
	}

	if len(matches) > 0 {
		m.logger.Logf("merged %d transfers into %d files", len(matches), len(files))
	}

	// Remove the directory if there are no files, otherwise setup an inner dir for the uploaded file.
	if len(files) == 0 {
		// delete the new directory as there's nothing to merge
		if err := os.RemoveAll(dir); err != nil {
			el.Add(err)
		}
	} else {
		dir = filepath.Join(dir, "uploaded")
		os.MkdirAll(dir, 0777)
	}

	// Write each file to our storage
	for i := range files {
		// Optionally Flatten Batches
		if m.cfg != nil && m.cfg.FlattenBatches != nil {
			fmt.Printf("attempting flatten: %#v\n", m.cfg)
			if file, err := files[i].FlattenBatches(); err != nil {
				el.Add(err)
			} else {
				files[i] = file
			}
		}
		// Write our file to the mergable directory
		if err := writeFile(dir, files[i]); err != nil {
			el.Add(fmt.Errorf("problem writing merged file: %v", err))
		}
		// Call our closure with the final file
		if err := f(files[i]); err != nil {
			el.Add(fmt.Errorf("problem from callback: %v", err))
		}
	}

	m.logger.Logf("wrote %d files", len(files))

	if !el.Empty() {
		return nil, el
	}

	return newProcessedTransfers(matches), nil
}

func writeFile(dir string, file *ach.File) error {
	var buf bytes.Buffer
	if err := ach.NewWriter(&buf).Write(file); err != nil {
		return fmt.Errorf("unable to buffer ACH file: %v", err)
	}
	filename := filepath.Join(dir, fmt.Sprintf("%s.ach", hash(buf.Bytes())))
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func hash(data []byte) string {
	ss := sha256.New()
	ss.Write(data)
	return hex.EncodeToString(ss.Sum(nil))
}
