package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultdata"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultsync"
)

const SiteLoginVaultStorageType = "site-login"

var (
	_ vaultsync.DataSyncer    = (*LoginVaultModel)(nil)
	_ vaultsync.StorageSyncer = (*LoginVaultStorage)(nil)
)

var siteLoginLastIndex uint32 = 0

func siteLoginLoadNextIndex() uint32 {
	return atomic.AddUint32(&siteLoginLastIndex, 1)
}

var LoginMetaDataSiteURLKey = "siteURL"

type LoginVaultModel struct {
	ID         uint32
	ExternalID string

	Data     []byte
	MetaData map[string]string

	Version    int  // for sync
	IsNew      bool // for sync
	IsUpdate   bool // for sync
	IsDelete   bool // for sync
	IsConflict bool // for sync
}

func (m *LoginVaultModel) GetID() uint32 {
	return m.ID
}

func (m *LoginVaultModel) GetVaultID() string {
	return m.ExternalID
}

func (m *LoginVaultModel) GetVersion() int {
	return m.Version
}

func (m *LoginVaultModel) GetIsNew() bool {
	return m.IsNew
}

func (m *LoginVaultModel) GetIsDelete() bool {
	return m.IsDelete
}

func (m *LoginVaultModel) GetIsUpdate() bool {
	return m.IsUpdate
}

func (m *LoginVaultModel) GetS3URL() string {
	return ""
}

func (m *LoginVaultModel) IsNeedSync() bool {
	return m.IsUpdate // TODO: check
}

func (m *LoginVaultModel) SetSite(siteURL string) {
	m.MetaData[LoginMetaDataSiteURLKey] = siteURL
}

func (m *LoginVaultModel) GetSite() string {
	siteURL, _ := m.MetaData[LoginMetaDataSiteURLKey]

	return siteURL
}

type siteLoginVaultStored struct {
	Data     []byte
	MetaData map[string]string
}

func siteLoginVaultStoredFromModel(model *LoginVaultModel) *siteLoginVaultStored {
	v := siteLoginVaultStored{
		Data:     model.Data,
		MetaData: model.MetaData,
	}

	return &v
}

func NewLoginVaultModel() *LoginVaultModel {
	m := LoginVaultModel{
		ID: siteLoginLoadNextIndex(),

		MetaData: make(map[string]string),

		IsNew:    true,
		IsUpdate: false,
	}

	return &m
}

type LoginSecreteData struct {
	Login    string
	Password string
}

type LoginVaultStorage struct {
	storage              map[uint32]*LoginVaultModel
	indexIDAndExternalID map[string]uint32

	crypt *vaultcrypt.VaultCrypt

	mux sync.RWMutex
}

func NewLoginVaultStorage(
	crypt *vaultcrypt.VaultCrypt,
) *LoginVaultStorage {
	s := LoginVaultStorage{
		crypt: crypt,

		storage:              make(map[uint32]*LoginVaultModel),
		indexIDAndExternalID: make(map[string]uint32),
	}

	return &s
}

type SiteLoginSavedStorage struct {
	Storage              map[uint32]*LoginVaultModel
	IndexIDAndExternalID map[string]uint32
}

func (s *LoginVaultStorage) LoadFromLocalFile(filePathDB string) error {
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

	savedStorage := SiteLoginSavedStorage{
		Storage:              make(map[uint32]*LoginVaultModel),
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

func (s *LoginVaultStorage) SaveToFile(filePathDB string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	file, err := os.OpenFile(filePathDB, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer file.Sync()
	defer file.Close()

	savedStorage := SiteLoginSavedStorage{
		Storage:              s.storage,
		IndexIDAndExternalID: s.indexIDAndExternalID,
	}

	err = gob.NewEncoder(file).Encode(&savedStorage)

	return err
}

func (s *LoginVaultStorage) GetKind() string {
	return SiteLoginVaultStorageType
}

func (s *LoginVaultStorage) LoadForSync() ([]vaultsync.DataSyncer, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	arr := make([]vaultsync.DataSyncer, 0, len(s.storage))

	for _, model := range s.storage {
		arr = append(arr, model)
	}

	return arr, nil
}

func (s *LoginVaultStorage) SetConflictFlag(id uint32) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	model, ok := s.storage[id]
	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	model.IsConflict = true

	return nil
}

func (s *LoginVaultStorage) SerializeToVault(data interface{}) ([]byte, error) {
	siteLoginModel, ok := data.(*LoginVaultModel)

	if !ok {
		return nil, ErrInvalidType
	}

	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(siteLoginVaultStoredFromModel(siteLoginModel))

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *LoginVaultStorage) DeserializeFromVault(dst []byte) (interface{}, error) {
	var vStored siteLoginVaultStored

	err := gob.NewDecoder(bytes.NewReader(dst)).Decode(&vStored)

	if err != nil {
		return nil, err
	}

	return &vStored, nil
}

func (s *LoginVaultStorage) UpdateAfterSyncByID(model vaultsync.DataSyncer, externalID string, version int) error {
	id := model.GetID()

	siteLoginModel, ok := s.storage[id]

	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	siteLoginModel.ExternalID = externalID
	siteLoginModel.Version = version
	siteLoginModel.IsNew = false
	s.indexIDAndExternalID[externalID] = id

	return nil
}

func (s *LoginVaultStorage) ConfirmDeleteAfterSyncByID(model vaultsync.DataSyncer) error {
	id := model.GetID()

	_, ok := s.storage[id]

	if !ok {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	delete(s.indexIDAndExternalID, model.GetVaultID())
	delete(s.storage, id)

	return nil
}

func (s *LoginVaultStorage) CreateDataStorage(externalID string, version int, data interface{}, _ string) error {
	vs, ok := data.(*siteLoginVaultStored)

	if !ok {
		return ErrInvalidType
	}

	_, ok = s.indexIDAndExternalID[externalID]

	if ok {
		// TODO: Logs or replace
		return nil
	}

	loginVaultModel := NewLoginVaultModel()

	loginVaultModel.Data = vs.Data
	loginVaultModel.MetaData = vs.MetaData
	loginVaultModel.Version = version
	loginVaultModel.ExternalID = externalID
	loginVaultModel.IsNew = false

	s.storage[loginVaultModel.ID] = loginVaultModel
	s.indexIDAndExternalID[externalID] = loginVaultModel.ID

	return nil
}

func (s *LoginVaultStorage) UpdateDataStorage(externalID string, version int, data interface{}) error {
	vs, ok := data.(*siteLoginVaultStored)

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
	model.Version = version

	return nil
}

func (s *LoginVaultStorage) DeleteDataStorage(externalID string, version int) error {
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

func (s *LoginVaultStorage) Create(data *LoginSecreteData, siteURL string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	newM := NewLoginVaultModel()

	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(data)

	if err != nil {
		return err
	}

	encryptedData, err := s.crypt.Encrypt(buffer.Bytes())
	if err != nil {
		return err
	}

	newM.Data = encryptedData
	newM.SetSite(siteURL)

	s.storage[newM.ID] = newM
	//s.indexIDAndExternalID[newM.ExternalID] = newM.ID

	return nil
}

func (s *LoginVaultStorage) GetAll() []*LoginVaultModel {
	s.mux.RLock()
	defer s.mux.RUnlock()

	arr := make([]*LoginVaultModel, 0, len(s.storage))

	for _, model := range s.storage {
		arr = append(arr, model)
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].ID < arr[j].ID
	})

	return arr
}

func (s *LoginVaultStorage) ViewDataByID(id uint32) (*LoginSecreteData, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	model, ok := s.storage[id]
	if !ok || model.IsDelete {
		return nil, vaultdata.ErrNotFoundVaultInStorage
	}

	decryptedData, err := s.crypt.Decrypt(model.Data)

	if err != nil {
		return nil, err
	}

	var data LoginSecreteData

	if err := gob.NewDecoder(bytes.NewReader(decryptedData)).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *LoginVaultStorage) DeleteByID(id uint32) error {
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

func (s *LoginVaultStorage) UpdateByID(id uint32, login, password string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	model, ok := s.storage[id]

	if !ok || model.IsDelete {
		return vaultdata.ErrNotFoundVaultInStorage
	}

	loginData := LoginSecreteData{
		Login:    login,
		Password: password,
	}

	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(loginData)

	if err != nil {
		return err
	}

	encrypted, err := s.crypt.Encrypt(buffer.Bytes())

	if err != nil {
		return err
	}

	model.Data = encrypted

	model.IsUpdate = false

	return nil
}
