package ethash

import "errors"

// FlashbotsAPI exposes Flashbots related methods for the RPC interface.
type FlashbotsAPI struct {
	ethash *Ethash
}

// GetWork returns a work package for external miner.
//
// The work package consists of 5 strings:
//   result[0] - 32 bytes hex encoded current block header pow-hash
//   result[1] - 32 bytes hex encoded seed hash used for DAG
//   result[2] - 32 bytes hex encoded boundary condition ("target"), 2^256/difficulty
//   result[3] - hex encoded block number
//   result[4], 32 bytes hex encoded parent block header pow-hash
//   result[5], hex encoded gas limit
//   result[6], hex encoded gas used
//   result[7], hex encoded transaction count
//   result[8], hex encoded uncle count
//   result[9], RLP encoded header with additonal empty extra data bytes
//   result[10] - hex encoded profit generated from this block
func (api *FlashbotsAPI) GetWork() ([11]string, error) {
	if api.ethash.remote == nil {
		return [11]string{}, errors.New("not supported")
	}

	var (
		workCh = make(chan [11]string, 1)
		errc   = make(chan error, 1)
	)
	select {
	case api.ethash.remote.fetchWorkCh <- &sealWork{errc: errc, res: workCh}:
	case <-api.ethash.remote.exitCh:
		return [11]string{}, errEthashStopped
	}
	select {
	case work := <-workCh:
		return work, nil
	case err := <-errc:
		return [11]string{}, err
	}
}
