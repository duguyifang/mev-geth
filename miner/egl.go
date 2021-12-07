package miner

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/miner/egl"
	"math/big"
	"time"
)

func (w *worker) eglLoop() {
	if w.config.EglAddress != "" {
		// Sleep for 15 seconds to allow ipc to start
		time.Sleep(15 * time.Second)
		log.Info("EGL initialized", "address", w.config.EglAddress)
		client, clientErr := ethclient.Dial(w.config.EglConnectUrl)
		eglInstance, instanceErr := egl.NewEgl(common.HexToAddress(w.config.EglAddress), client)

		if clientErr != nil {
			log.Error("ETH client dial error", "message", clientErr, "url", w.config.EglConnectUrl)
		}

		if instanceErr != nil {
			log.Error("ETH contract instance error", "message", instanceErr, "egladdress", w.config.EglAddress)
		}
		for {
			rawDesiredEgl, contractReadErr := eglInstance.DesiredEgl(nil)
			var desiredEgl, _ = new(big.Int).SetString(rawDesiredEgl.String(), 10)
			w.config.GasFloor = desiredEgl.Uint64() / 2
			w.config.GasCeil = desiredEgl.Uint64()
			log.Info("EGL gas limit adjustment", "desiredegl", desiredEgl, "gastarget", w.config.GasFloor, "gaslimit", w.config.GasCeil)

			if contractReadErr != nil {
				log.Error("EGL contract read error", "message", contractReadErr)
			}
			time.Sleep(w.config.EglInterval)
		}
	} else {
		log.Info("EGL not initialized", "reason", "No EGL contract address configured")
		return
	}
}
