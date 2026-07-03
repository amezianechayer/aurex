package sqlite

import "github.com/amezianechayer/corren/storage"

type Driver struct{}

func (d *Driver) Initialize() error {
	return nil
}

func (d *Driver) NewStore(name string) (storage.Store, error) {
	return NewStore(name)
}

func init() {
	storage.RegisterDriver("sqlite", &Driver{})
}
