package minermanage

import (
	"encoding/json"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/prometheus/common/log"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-miner/node/modules/dtypes"
)

const actorKey = "miner-actors"
const defaultKey = "default-actor"

var ErrNoDefault = xerrors.Errorf("not set default key")

type MinerManageAPI interface {
	Put(addr dtypes.MinerInfo) error
	Set(addr dtypes.MinerInfo) error
	Has(checkAddr address.Address) bool
	Get(checkAddr address.Address) *dtypes.MinerInfo
	List() ([]dtypes.MinerInfo, error)
	Remove(addrs []address.Address) error
	Count() int
}

type MinerManager struct {
	miners []dtypes.MinerInfo

	da dtypes.MetadataDS
	lk sync.Mutex
}

func NewMinerManger(ds dtypes.MetadataDS) (*MinerManager, error) {
	addrBytes, err := ds.Get(datastore.NewKey(actorKey))
	if err != nil && err != datastore.ErrNotFound {
		return nil, err
	}

	var miners []dtypes.MinerInfo

	if err != datastore.ErrNotFound {
		err = json.Unmarshal(addrBytes, &miners)
		if err != nil {
			return nil, err
		}
	}

	return &MinerManager{da: ds, miners: miners}, nil
}

func (m *MinerManager) Put(miner dtypes.MinerInfo) error {
	m.lk.Lock()
	defer m.lk.Unlock()

	if m.Has(miner.Addr) {
		log.Warnf("addr %s has exit", miner.Addr)
		return nil
	}

	newMiner := append(m.miners, miner)
	addrBytes, err := json.Marshal(newMiner)
	if err != nil {
		return err
	}
	err = m.da.Put(datastore.NewKey(actorKey), addrBytes)
	if err != nil {
		return err
	}

	m.miners = newMiner
	return nil
}

func (m *MinerManager) Set(miner dtypes.MinerInfo) error {
	m.lk.Lock()
	defer m.lk.Unlock()

	for k, addr := range m.miners {
		if addr.Addr.String() == miner.Addr.String() {
			if miner.Sealer.ListenAPI != "" && miner.Sealer.ListenAPI != m.miners[k].Sealer.ListenAPI {
				m.miners[k].Sealer.ListenAPI = miner.Sealer.ListenAPI
			}

			if miner.Sealer.Token != "" && miner.Sealer.Token != m.miners[k].Sealer.Token {
				m.miners[k].Sealer.Token = miner.Sealer.Token
			}

			if miner.Wallet.ListenAPI != "" && miner.Wallet.ListenAPI != m.miners[k].Wallet.ListenAPI {
				m.miners[k].Wallet.ListenAPI = miner.Wallet.ListenAPI
			}

			if miner.Wallet.Token != "" && miner.Wallet.Token != m.miners[k].Wallet.Token {
				m.miners[k].Wallet.Token = miner.Wallet.Token
			}

			addrBytes, err := json.Marshal(m.miners)
			if err != nil {
				return err
			}

			err = m.da.Put(datastore.NewKey(actorKey), addrBytes)
			if err != nil {
				return err
			}

			break
		}
	}

	return nil
}

func (m *MinerManager) Has(addr address.Address) bool {
	for _, miner := range m.miners {
		if miner.Addr.String() == addr.String() {
			return true
		}
	}

	return false
}

func (m *MinerManager) Get(addr address.Address) *dtypes.MinerInfo {
	m.lk.Lock()
	defer m.lk.Unlock()

	for k := range m.miners {
		if m.miners[k].Addr.String() == addr.String() {
			return &m.miners[k]
		}
	}

	return nil
}

func (m *MinerManager) List() ([]dtypes.MinerInfo, error) {
	m.lk.Lock()
	defer m.lk.Unlock()

	return m.miners, nil
}

func findAddress(addr address.Address, addrs []address.Address) bool {
	for _, a := range addrs {
		if a.String() != addr.String() {
			return true
		}
	}

	return false
}

func (m *MinerManager) Remove(addrs []address.Address) error {
	m.lk.Lock()
	defer m.lk.Unlock()

	var newMiners []dtypes.MinerInfo
	for _, miner := range m.miners {
		if !findAddress(miner.Addr, addrs) {
			newMiners = append(newMiners, miner)
		}
	}

	addrBytes, err := json.Marshal(newMiners)
	if err != nil {
		return err
	}
	err = m.da.Put(datastore.NewKey(actorKey), addrBytes)
	if err != nil {
		return err
	}

	m.miners = newMiners

	return nil
}

func (m *MinerManager) SetDefault(addr address.Address) error {
	return m.da.Put(datastore.NewKey(defaultKey), addr.Bytes())
}

func (m *MinerManager) Default() (address.Address, error) {
	bytes, err := m.da.Get(datastore.NewKey(defaultKey))
	if err != nil {
		// set the address with index 0 as the default address
		if len(m.miners) == 0 {
			return address.Undef, ErrNoDefault
		}

		err = m.SetDefault(m.miners[0].Addr)
		if err != nil {
			return address.Undef, err
		}

		return m.miners[0].Addr, nil
	}

	return address.NewFromBytes(bytes)
}

func (m *MinerManager) Count() int {
	m.lk.Lock()
	defer m.lk.Unlock()

	return len(m.miners)
}

var _ MinerManageAPI = &MinerManager{}
