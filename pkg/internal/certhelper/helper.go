package certhelper

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"hash/fnv"
	"io"
	"os"
	"sync"
	"time"
)

type CertificateWatcher interface {
	Start() error
	Stop() error
}

type Callback = func(path string) error

type fsnotifyCertificateWatcher struct {
	log           *logrus.Entry
	once          sync.Once
	cancelChannel chan chan error
	watcher       *fsnotify.Watcher
	callback      Callback
}

func NewFileCertificateHelper(log *logrus.Entry, paths []string, callback Callback) (CertificateWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher for the root certificate")
	}

	for _, path := range paths {
		if err = watcher.Add(path); err != nil {
			return nil, errors.Wrap(err, "failed to add certificate path to watcher")
		}
	}

	return &fsnotifyCertificateWatcher{
		log:           log,
		watcher:       watcher,
		cancelChannel: make(chan chan error),
		callback:      callback,
	}, nil
}

func (f *fsnotifyCertificateWatcher) Start() error {
	f.once.Do(func() {
		go f.backgroundWorker()
	})

	return nil
}

func (f *fsnotifyCertificateWatcher) backgroundWorker() {
	var cancelChannel chan error
	var err error
	defer func() {
		if cancelChannel != nil {
			cancelChannel <- err
		}
	}()

	lastChange := time.Now().Add(-1 * time.Minute)

	for {
		select {
		case err = <-f.watcher.Errors:

		case event := <-f.watcher.Events:
			if time.Now().Add(-1 * time.Minute).Before(lastChange) {
				continue
			}

			if err = f.callback(event.Name); err != nil {

			}

			lastChange = time.Now()
		case cancelChannel = <-f.cancelChannel:
			return
		}
	}
}

func (f *fsnotifyCertificateWatcher) Stop() error {
	callback := make(chan error)
	f.cancelChannel <- callback

	return <-callback
}

func NewManualCertificateWatcher(log *logrus.Entry, paths []string, callback Callback) (CertificateWatcher, error) {
	return &manualCertificateWatcher{
		log:           log,
		once:          sync.Once{},
		cancelChannel: make(chan chan error),
		callback:      callback,
		paths:         paths,
		hashes:        map[string]uint64{},
	}, nil
}

type manualCertificateWatcher struct {
	log           *logrus.Entry
	once          sync.Once
	cancelChannel chan chan error
	callback      Callback
	paths         []string
	hashes        map[string]uint64
}

func (f *manualCertificateWatcher) Start() error {
	f.once.Do(func() {
		go f.backgroundWorker()
	})

	return nil
}

func (f *manualCertificateWatcher) backgroundWorker() {
	var cancelChannel chan error
	var err error
	defer func() {
		if cancelChannel != nil {
			cancelChannel <- err
		}
	}()

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			if f.hasChanged() {
				err = f.callback("")
			}
		case cancelChannel = <-f.cancelChannel:
			return
		}
	}
}

func (f *manualCertificateWatcher) Stop() error {
	callback := make(chan error)
	f.cancelChannel <- callback

	return <-callback
}

func (f *manualCertificateWatcher) hasChanged() bool {
	var changeSeen bool
	hit := map[string]uint64{}

	for _, path := range f.paths {
		log := f.log.WithField("path", path)

		stat, err := os.Stat(path)
		if err != nil {
			log.WithError(err).Warn("failed to get stat for path")
			continue
		}

		hash, err := f.hashItem(stat)
		if err != nil {
			log.WithError(err).Warn("failed to hash item")
			continue
		}

		hit[stat.Name()] = hash

		if current, ok := f.hashes[stat.Name()]; ok {
			if current != hash {
				changeSeen = true
			}
		} else {
			changeSeen = true
		}
	}

	f.hashes = hit

	return changeSeen
}

func (f *manualCertificateWatcher) hashItem(stat os.FileInfo) (uint64, error) {
	if stat.IsDir() {
		return f.hashDirectory(stat)
	}

	return f.hashFile(stat)
}

func (f *manualCertificateWatcher) hashDirectory(stat os.FileInfo) (uint64, error) {
	hash := fnv.New64()
	entries, err := os.ReadDir(stat.Name())
	if err != nil {
		return 0, errors.Wrap(err, "failed to read directory")
	}

	for _, entry := range entries {
		if _, err = hash.Write([]byte(entry.Name())); err != nil {
			return 0, errors.Wrap(err, "failed to write name to hash for directory")
		}
	}

	return hash.Sum64(), nil
}

func (f *manualCertificateWatcher) hashFile(stat os.FileInfo) (uint64, error) {
	hash := fnv.New64()

	file, err := os.Open(stat.Name())
	if err != nil {
		return 0, errors.Wrap(err, "failed to open file to hash")
	}

	_, err = io.Copy(hash, file)
	if err != nil {
		return 0, errors.Wrap(err, "failed to write file to hash")
	}

	defer file.Close()

	return hash.Sum64(), nil
}
