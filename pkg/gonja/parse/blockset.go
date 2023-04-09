package parse

import (
	"fmt"
)

// BlockSet is a map of block names to block implementations.
type BlockSet map[string]*WrapperNode

// Exists indicates if the given block is already registered.
func (bs BlockSet) Exists(name string) bool {
	_, existing := bs[name]
	return existing
}

// Register registers a new block. An error will be returned, if there's already
// a block with the same name registered.
func (bs *BlockSet) Register(name string, w *WrapperNode) error {
	if bs.Exists(name) {
		return fmt.Errorf("block with name '%s' is already registered", name)
	}
	(*bs)[name] = w
	return nil
}

// Replace replaces an already registered block with a new implementation.
func (bs *BlockSet) Replace(name string, w *WrapperNode) error {
	if !bs.Exists(name) {
		return fmt.Errorf("block with name '%s' does not exist (therefore cannot be overridden)", name)
	}
	(*bs)[name] = w
	return nil
}
