package miner

import (
	"github.com/ethereum/go-ethereum/cmd/abigen/egl"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"time"
)

func (w *worker) eglLoop() {
	go func(eglAddress string, gastarget uint64, gaslimit uint64) {
		log.Info("EGL Initialized", "address", eglAddress)
		//dialUrl := "http://localhost:7545"
		dialUrl := "https://ropsten.infura.io/v3/"
		client, clientErr := ethclient.Dial(dialUrl)
		eglInstance, instanceErr := egl.NewEgl(common.HexToAddress(eglAddress), client)

		if clientErr != nil {
			log.Error("ETH client dial error", "message", clientErr, "url", dialUrl)
		}

		if instanceErr != nil {
			log.Error("ETH client dial error", "message", instanceErr, "egladdress", eglAddress)
		}

		for {
			rawDesiredEgl, err := eglInstance.DesiredEgl(nil)
			log.Info("EGL gas limit adjustment", "desiredegl", rawDesiredEgl, "gastarget", gastarget, "gaslimit", gaslimit)
			var desiredEgl, _ = new(big.Int).SetString(rawDesiredEgl.String(), 10)
			gastarget = desiredEgl.Uint64()
			gaslimit = desiredEgl.Uint64()
			if err != nil {
				log.Error("EGL error", "message", err)
			}
			time.Sleep(10 * time.Second)
		}
	}(w.config.EglAddress, w.config.GasFloor, w.config.GasCeil)
}
