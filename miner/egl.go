package miner

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/miner/egl"
	"math/big"
	"time"
)

type EglConfig struct {
	enableCh     chan bool     // Should following EGL
	enabled      bool          // Should following EGL
	baseGasFloor uint64        // Original gas floor value
	baseGasCeil  uint64        // Original gas ceil value
	Address      string        // Address of EGL smart contract.
	ConnectUrl   string        // URL for the client to get desired EGL value from the contract.
	Interval     time.Duration // How often to check EGL value.
}

func (w *worker) eglLoop() {
	eglConfig := &w.config.Egl
	eglConfig.enableCh = make(chan bool, 1)
	eglConfig.enabled = false
	eglConfig.baseGasCeil = w.config.GasCeil
	eglConfig.baseGasFloor = w.config.GasFloor
	log.Info("EGL Base Gas Values", "gasfloor", eglConfig.baseGasFloor, "gasceil", eglConfig.baseGasCeil)

	// Enable by default
	w.enableEgl()

	if eglConfig.Address != "" {
		time.Sleep(5 * time.Second)
		log.Info("EGL initialized", "address", eglConfig.Address)

		stopEgl := make(chan bool, 1)
		for {
			select {
			case enabled := <-eglConfig.enableCh:
				eglConfig.enabled = enabled
				log.Info("EGL Status", "enabled", enabled, "gasfloor", w.config.GasFloor, "gasceil", w.config.GasCeil)
				if enabled {
					go eglRunner(eglConfig, w.config, stopEgl)
				} else {
					stopEgl <- true
					w.config.GasFloor = eglConfig.baseGasFloor
					w.config.GasCeil = eglConfig.baseGasCeil
					log.Info("EGL Disabled. Base Gas Limit Values Restored", "gasfloor", w.config.GasFloor, "gasceil", w.config.GasCeil)
				}
			}
		}
	} else {
		log.Warn("EGL not initialized", "reason", "No EGL contract address configured")
		return
	}
}

func eglRunner(eglConfig *EglConfig, workerConfig *Config, stopEgl chan bool) {
	eglClient, eglClientErr := ethclient.Dial(eglConfig.ConnectUrl)
	eglContractInstance, eglContractErr := egl.NewEgl(common.HexToAddress(eglConfig.Address), eglClient)

	if eglClientErr != nil {
		log.Error("ETH Client Dial Error", "message", eglClientErr, "url", eglConfig.ConnectUrl)
		return
	}

	if eglContractErr != nil {
		log.Error("ETH Contract Instance Error", "message", eglContractErr, "egladdress", eglConfig.Address)
		return
	}

	for {
		select {
		case <-stopEgl:
			log.Info("EGL Stopped")
			return
		default:
			if eglConfig.enabled {
				desiredEgl := getDesiredEgl(eglContractInstance)
				if desiredEgl != nil {
					workerConfig.GasFloor = desiredEgl.Uint64() / 2
					workerConfig.GasCeil = desiredEgl.Uint64()
					log.Info("EGL adjusted gas limits", "desiredegl", desiredEgl, "gastarget", workerConfig.GasFloor, "gaslimit", workerConfig.GasCeil)
				}
				time.Sleep(eglConfig.Interval)
			}
		}
	}
}

func getDesiredEgl(eglContract *egl.Egl) *big.Int {
	rawDesiredEgl, contractReadErr := eglContract.DesiredEgl(nil)
	if contractReadErr != nil {
		log.Error("EGL contract read error", "message", contractReadErr)
		return nil
	}
	var desiredEgl, _ = new(big.Int).SetString(rawDesiredEgl.String(), 10)
	return desiredEgl
}

// enableEgl enables following desired EGL
func (w *worker) enableEgl() {
	if w.config.Egl.enabled {
		log.Warn("EGL Already Enabled", "status", "ignoring...")
		return
	}
	w.config.Egl.enableCh <- true
}

// disableEgl enables following desired EGL
func (w *worker) disableEgl() {
	if !w.config.Egl.enabled {
		log.Warn("EGL Already Disabled", "status", "ignoring...")
		return
	}
	w.config.Egl.enableCh <- false
}
