package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jaevor/go-nanoid"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultdata"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultsync"
	"golang.org/x/sync/errgroup"
)

const FileVaultStorageType = "file"

var randID, _ = nanoid.Standard(36)

var (
	_ vaultsync.DataSyncer    = (*FileVaultModel)(nil)
	_ vaultsync.StorageSyncer = (*FileVaultStorage)(nil)
)

var fileLastIndex uint32 = 0

func fileLoadNextIndex() uint32 {
	return atomic.AddUint32(&fileLastIndex, 1)
}

var FileMetaDataNameKey = "file-name"
var FileMetaDataExtensionKey = "extension"
var FileMetaDataEncryptedKey = "encrypted-name"

type FileVaultModel struct {
	ID         uint32
	ExternalID string

	Data     []byte
	MetaData map[string]string
	S3URL    string

	Version    int  // for sync
	IsNew      bool // for sync
	IsUpdate   bool // for sync
	IsDelete   bool // for sync
	IsConflict bool // for sync
}

func (m *FileVaultModel) GetID() uint32 {
	return m.ID
}

func (m *FileVaultModel) GetVaultID() string {
	return m.ExternalID
}

func (m *FileVaultModel) GetVersion() int {
	return m.Version
}

func (m *FileVaultModel) GetIsNew() bool {
	return m.IsNew
}

func (m *FileVaultModel) GetIsDelete() bool {
	return m.IsDelete
}

func (m *FileVaultModel) GetIsUpdate() bool {
	return m.IsUpdate
}

func (m *FileVaultModel) GetS3URL() string {
	return m.S3URL
}

func (m *FileVaultModel) IsNeedSync() bool {
	return m.IsUpdate || m.IsDelete || m.IsNew // TODO: check
}

func (m *FileVaultModel) GetFileName() string {
	name, _ := m.MetaData[FileMetaDataNameKey]

	return name
}

func (m *FileVaultModel) SetFileName(name string) {
	m.MetaData[FileMetaDataNameKey] = name
}

func (m *FileVaultModel) GetExtensionName() string {
	extension, _ := m.MetaData[FileMetaDataExtensionKey]

	return extension
}

func (m *FileVaultModel) SetExtensionName(extension string) {
	m.MetaData[FileMetaDataExtensionKey] = extension
}

func (m *FileVaultModel) SetEncryptedName(encryptedName string) {
	m.MetaData[FileMetaDataEncryptedKey] = encryptedName
}

func (m *FileVaultModel) GetEncryptedName() string {
	encryptedName, _ := m.MetaData[FileMetaDataEncryptedKey]

	return encryptedName
}

type fileVaultStored struct {
	Data     []byte
	MetaData map[string]string
}

func fileVaultStoredFromModel(model *FileVaultModel) *fileVaultStored {
	v := fileVaultStored{
		Data:     model.Data,
		MetaData: model.MetaData,
	}

	return &v
}

func NewFileVaultModel() *FileVaultModel {
	m := FileVaultModel{
		ID: fileLoadNextIndex(),

		MetaData: make(map[string]string),

		IsNew:    true,
		IsUpdate: false,
	}

	return &m
}

type FileSecreteData struct {
	Key []byte
}

type FileVaultStorage struct {
	storage              map[uint32]*FileVaultModel
	indexIDAndExternalID map[string]uint32

	vclient *vaultclient.Client
	crypt   *vaultcrypt.VaultCrypt

	mux sync.RWMutex
}

func NewFileVaultStorage(
	crypt *vaultcrypt.VaultCrypt,
	vclient *vaultclient.Client,
) *FileVaultStorage {
	s := FileVaultStorage{
		crypt:   crypt,
		vclient: vclient,

		storage:              make(map[uint32]*FileVaultModel),
		indexIDAndExternalID: make(map[string]uint32),
	}

	return &s
}

type FileSavedStorage struct {
	Storage              map[uint32]*FileVaultModel
	IndexIDAndExternalID map[string]uint32
}

func (s *FileVaultStorage) LoadFromLocalFile(filePathDB string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	var file *os.File
	var err error

	file, err = os.Open(filePathDB)
	defer file.Close()

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		file, err = os.Create(filePathDB)

		if err != nil {
			return err
		}
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	savedStorage := FileSavedStorage{
		Storage:              make(map[uint32]*FileVaultModel),
		IndexIDAndExternalID: make(map[string]uint32),
	}

	if fileInfo.Size() == 0 {
		return nil
	}

	err = gob.NewDecoder(file).Decode(&savedStorage)
	if err != nil {
		return err
	}

	s.storage = savedStorage.Storage
	s.indexIDAndExternalID = savedStorage.IndexIDAndExternalID

	return nil
}

func (s *FileVaultStorage) SaveToFile(filePathDB string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	file, err := os.OpenFile(filePathDB, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer file.Sync()
	defer file.Close()

	savedStorage := FileSavedStorage{
		Storage:              s.storage,
		IndexIDAndExternalID: s.indexIDAndExternalID,
	}

	err = gob.NewEncoder(file).Encode(&savedStorage)

	return err
}

func (s *FileVaultStorage) GetKind() string {
	return FileVaultStorageType
}

func (s *FileVaultStorage) LoadForSync() ([]vaultsync.DataSyncer, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	arr := make([]vaultsync.DataSyncer, 0, len(s.storage))

	for _, model := range s.storage {
		arr = append(arr, model)
	}

	return arr, nil
}

func (s *FileVaultStorage) SetConflictFlag(id uint32) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	model, ok := s.storage[id]
	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	model.IsConflict = true

	return nil
}

func (s *FileVaultStorage) SerializeToVault(data interface{}) ([]byte, error) {
	fileVaultModel, ok := data.(*FileVaultModel)

	if !ok {
		return nil, ErrInvalidType
	}

	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(fileVaultStoredFromModel(fileVaultModel))

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *FileVaultStorage) DeserializeFromVault(dst []byte) (interface{}, error) {
	var vStored fileVaultStored

	err := gob.NewDecoder(bytes.NewReader(dst)).Decode(&vStored)

	if err != nil {
		return nil, err
	}

	return &vStored, nil
}

func (s *FileVaultStorage) UpdateAfterSyncByID(model vaultsync.DataSyncer, externalID string, version int) error {
	id := model.GetID()

	siteLoginModel, ok := s.storage[id]

	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	siteLoginModel.ExternalID = externalID
	siteLoginModel.Version = version
	s.indexIDAndExternalID[externalID] = id

	return nil
}

func (s *FileVaultStorage) ConfirmDeleteAfterSyncByID(model vaultsync.DataSyncer) error {
	id := model.GetID()

	_, ok := s.storage[id]

	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	delete(s.indexIDAndExternalID, model.GetVaultID())
	delete(s.storage, id)

	return nil
}

func (s *FileVaultStorage) CreateDataStorage(externalID string, version int, data interface{}, s3URL string) error {
	vs, ok := data.(*fileVaultStored)

	if !ok {
		return ErrInvalidType
	}

	_, ok = s.indexIDAndExternalID[externalID]

	if ok {
		// TODO: Logs or replace
		return nil
	}

	fileVaultModel := NewFileVaultModel()

	fileVaultModel.Data = vs.Data
	fileVaultModel.MetaData = vs.MetaData
	fileVaultModel.Version = version
	fileVaultModel.ExternalID = externalID
	fileVaultModel.IsNew = false

	if s3URL != "" {
		fileVaultModel.S3URL = s3URL
	}

	s.storage[fileVaultModel.ID] = fileVaultModel
	s.indexIDAndExternalID[externalID] = fileVaultModel.ID

	return nil
}

func (s *FileVaultStorage) UpdateDataStorage(externalID string, version int, data interface{}) error {
	vs, ok := data.(*fileVaultStored)

	if !ok {
		return ErrInvalidType
	}

	id, ok := s.indexIDAndExternalID[externalID]

	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	model, ok := s.storage[id]

	if !ok {
		delete(s.indexIDAndExternalID, externalID)
		// TODO: logs
		return vaultdata.ErrNotFoundVaultInStorage
	}

	if model.IsNeedSync() {
		// TODO: Logs or replace
		return nil
	}

	if model.Version > version {
		// TODO: Logs or replace
		return nil
	}

	model.Data = vs.Data
	model.MetaData = vs.MetaData
	model.IsNew = false
	model.Version = version

	return nil
}

func (s *FileVaultStorage) DeleteDataStorage(externalID string, version int) error {
	id, ok := s.indexIDAndExternalID[externalID]

	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	model, ok := s.storage[id]

	if !ok {
		delete(s.indexIDAndExternalID, externalID)
		// TODO: Need logs
		return vaultdata.ErrNotFoundVaultInStorage
	}

	if model.IsNeedSync() {
		// TODO: Logs or replace
		return nil
	}

	if model.Version > version {
		// TODO: Logs or replace
		return nil
	}

	delete(s.indexIDAndExternalID, externalID)
	delete(s.storage, id)

	return nil
}

// For storage!

func (s *FileVaultStorage) GetAll() []*FileVaultModel {
	s.mux.RLock()
	defer s.mux.RUnlock()

	arr := make([]*FileVaultModel, 0, len(s.storage))

	for _, model := range s.storage {
		arr = append(arr, model)
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].ID < arr[j].ID
	})

	return arr
}

func (s *FileVaultStorage) UploadFile(ctx context.Context, file *os.File) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	newM := NewFileVaultModel()
	newM.SetFileName(fileInfo.Name())
	newM.SetExtensionName(filepath.Ext(fileInfo.Name()))

	encryptedName := randID()
	newM.SetEncryptedName(encryptedName)

	encryptedKey := make([]byte, 256)
	_, err = rand.Read(encryptedKey)
	if err != nil {
		return err
	}

	newSecretData := FileSecreteData{
		Key: encryptedKey,
	}

	var buffer bytes.Buffer
	err = gob.NewEncoder(&buffer).Encode(&newSecretData)

	if err != nil {
		return err
	}

	encryptedData, err := s.crypt.Encrypt(buffer.Bytes())

	if err != nil {
		return err
	}

	newM.Data = encryptedData

	reader, writer := io.Pipe()
	defer reader.Close()

	w, err := s.crypt.EncryptStream(writer, encryptedKey)

	if err != nil {
		return err
	}

	g := new(errgroup.Group)

	g.Go(func() error {
		defer writer.Close()
		_, err = io.Copy(w, file)

		return err
	})

	result, err := s.vclient.VaultUpload(ctx, reader)

	if err != nil {
		return err
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	newM.S3URL = result

	s.storage[newM.ID] = newM

	return nil
}

func (s *FileVaultStorage) DownloadFile(ctx context.Context, id uint32, filePath string) error {
	s.mux.RLock()
	defer s.mux.RUnlock()

	model, ok := s.storage[id]

	if !ok || model.IsDelete {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	decryptedData, err := s.crypt.Decrypt(model.Data)

	if err != nil {
		return err
	}

	var data FileSecreteData

	err = gob.NewDecoder(bytes.NewReader(decryptedData)).Decode(&data)
	if err != nil {
		return err
	}

	if model.S3URL == "" || len(data.Key) == 0 {
		return errors.New("invalid data")
	}

	file, err := os.Create(filepath.Join(filePath, model.GetFileName()))

	if err != nil {
		return err
	}

	defer file.Sync()
	defer file.Close()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	body, err := s.vclient.VaultDownload(ctxWithTimeout, model.S3URL)
	if err != nil {
		return err
	}

	defer body.Close()

	r, err := s.crypt.DecryptStream(body, data.Key)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, r)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileVaultStorage) DeleteFile(_ context.Context, id uint32) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	model, ok := s.storage[id]

	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	model.IsUpdate = false
	model.IsDelete = !model.IsDelete

	return nil
}
