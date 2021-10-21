// (c) 2021, Flare Networks Limited. All rights reserved.
// Please see the file LICENSE for licensing terms.

package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	flareChainID                  = new(big.Int).SetUint64(14) // https://github.com/ethereum-lists/chains/blob/master/_data/chains/eip155-14.json
	songbirdChainID               = new(big.Int).SetUint64(19) // https://github.com/ethereum-lists/chains/blob/master/_data/chains/eip155-19.json
	costonChainID                 = new(big.Int).SetUint64(16) // https://github.com/ethereum-lists/chains/blob/master/_data/chains/eip155-16.json
	stateConnectorActivationTime  = new(big.Int).SetUint64(1636070400)
	stateConnectorFinalUpdateTime = new(big.Int).SetUint64(1635984000)
	tr                            = &http.Transport{
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     60 * time.Second,
		DisableCompression:  true,
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
	apiRetries    = 3
	apiRetryDelay = 1 * time.Second
)

func GetStateConnectorActivated(chainID *big.Int, blockTime *big.Int) bool {
	// Return true if chainID is 16 or if block.timestamp is greater than the state connector activation time on any chain
	return chainID.Cmp(costonChainID) == 0 || blockTime.Cmp(stateConnectorActivationTime) > 0
}

func GetStateConnectorGasDivisor(chainID *big.Int, blockTime *big.Int) uint64 {
	switch {
	default:
		return 3
	}
}

func GetMaxAllowedChains(chainID *big.Int, blockTime *big.Int) uint64 {
	switch {
	default:
		return 5
	}
}

func GetStateConnectorContractAddr(chainID *big.Int, blockTime *big.Int) string {
	switch {
	case chainID.Cmp(songbirdChainID) == 0:
		switch {
		case blockTime.Cmp(stateConnectorFinalUpdateTime) < 0:
			return "0x1000000000000000000000000000000000000001"
		default:
			return "0x1000000000000000000000000000000000000001"
		}
	default:
		return "0x1000000000000000000000000000000000000001"
	}
}

func SetPaymentFinalitySelector(chainID *big.Int, blockTime *big.Int) []byte {
	switch {
	default:
		return []byte{0xb3, 0x8f, 0x3d, 0x3e}
	}
}

// =======================================================
// Proof of Work Common
// =======================================================

type GetPoWRequestPayload struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type GetPoWBlockHeaderResult struct {
	Hash          string `json:"hash"`
	Confirmations uint64 `json:"confirmations"`
	Height        uint64 `json:"height"`
}

type GetPoWBlockHeaderResp struct {
	Result GetPoWBlockHeaderResult `json:"result"`
	Error  interface{}             `json:"error"`
}

func GetPoWBlockHeader(ledgerHash string, requiredConfirmations uint64, chainURL string, username string, password string) (uint64, bool) {
	data := GetPoWRequestPayload{
		Method: "getblockheader",
		Params: []string{
			ledgerHash,
		},
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return 0, true
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", chainURL, body)
	if err != nil {
		return 0, true
	}
	req.Header.Set("Content-Type", "application/json")
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, true
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 0, true
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, true
	}
	var jsonResp GetPoWBlockHeaderResp
	err = json.Unmarshal(respBody, &jsonResp)
	if err != nil {
		return 0, true
	}
	if jsonResp.Error != nil {
		return 0, false
	} else if jsonResp.Result.Confirmations < requiredConfirmations {
		return 0, false
	}
	return jsonResp.Result.Height, false
}

type GetPoWTxRequestParams struct {
	TxID    string `json:"txid"`
	Verbose bool   `json:"verbose"`
}

type GetPoWTxRequestPayload struct {
	Method string                `json:"method"`
	Params GetPoWTxRequestParams `json:"params"`
}

type GetPoWTxResult struct {
	TxID          string `json:"txid"`
	BlockHash     string `json:"blockhash"`
	Confirmations uint64 `json:"confirmations"`
	Vout          []struct {
		Value        float64 `json:"value"`
		N            uint64  `json:"n"`
		ScriptPubKey struct {
			Type      string   `json:"type"`
			Addresses []string `json:"addresses"`
		} `json:"scriptPubKey"`
	} `json:"vout"`
}

type GetPoWTxResp struct {
	Result GetPoWTxResult `json:"result"`
	Error  interface{}    `json:"error"`
}

func GetPoWTx(instructions Instructions, basicAuth BasicAuth) ([]byte, uint64, bool) {
	voutN := instructions.Utxo
	data := GetPoWTxRequestPayload{
		Method: "getrawtransaction",
		Params: GetPoWTxRequestParams{
			TxID:    hex.EncodeToString(instructions.TxId),
			Verbose: true,
		},
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return []byte{}, 0, true
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", basicAuth.ChainURL, body)
	if err != nil {
		return []byte{}, 0, true
	}
	req.Header.Set("Content-Type", "application/json")
	if basicAuth.Username != "" && basicAuth.Password != "" {
		req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, 0, true
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return []byte{}, 0, true
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, true
	}
	var jsonResp GetPoWTxResp
	err = json.Unmarshal(respBody, &jsonResp)
	if err != nil {
		return []byte{}, 0, true
	}
	if jsonResp.Error != nil {
		return []byte{}, 0, true
	}
	if uint64(len(jsonResp.Result.Vout)) <= uint64(voutN) {
		return []byte{}, 0, false
	}
	if jsonResp.Result.Vout[voutN].ScriptPubKey.Type != "pubkeyhash" || len(jsonResp.Result.Vout[voutN].ScriptPubKey.Addresses) != 1 {
		return []byte{}, 0, false
	}
	inBlock, getBlockErr := GetPoWBlockHeader(jsonResp.Result.BlockHash, jsonResp.Result.Confirmations, basicAuth.ChainURL, basicAuth.Username, basicAuth.Password)
	if getBlockErr {
		return []byte{}, 0, true
	}
	if inBlock == 0 || inBlock >= instructions.AvailableLedger {
		return []byte{}, 0, false
	}
	salt := crypto.Keccak256([]byte("FlareStateConnector_PAYMENTHASH"))
	destinationHash := crypto.Keccak256([]byte(jsonResp.Result.Vout[voutN].ScriptPubKey.Addresses[0]))
	dataHash := crypto.Keccak256(common.LeftPadBytes(common.FromHex(hexutil.EncodeUint64(instructions.ChainId)), 32))
	amountHash := crypto.Keccak256(common.LeftPadBytes(common.FromHex(hexutil.EncodeUint64(uint64(jsonResp.Result.Vout[voutN].Value*math.Pow(10, 8)))), 32))
	return crypto.Keccak256(salt, destinationHash, dataHash, amountHash), inBlock, false
}

func ProvePaymentPOW(instructions Instructions, basicAuth BasicAuth) (bool, bool) {
	paymentHash, inBlock, getPoWTxErr := GetPoWTx(instructions, basicAuth)
	if getPoWTxErr {
		return false, true
	}
	if len(paymentHash) > 0 && bytes.Equal(paymentHash, instructions.PaymentHash) && inBlock == instructions.Ledger {
		return true, false
	}
	return false, false
}

// =======================================================
// XRP
// =======================================================

type CheckXRPErrorResponse struct {
	Error string `json:"error"`
}

type GetXRPTxRequestParams struct {
	Transaction string `json:"transaction"`
	Binary      bool   `json:"binary"`
}

type GetXRPTxRequestPayload struct {
	Method string                  `json:"method"`
	Params []GetXRPTxRequestParams `json:"params"`
}

type GetXRPTxResponse struct {
	Destination     string `json:"Destination"`
	DestinationTag  int    `json:"DestinationTag"`
	TransactionType string `json:"TransactionType"`
	Hash            string `json:"hash"`
	InLedger        int    `json:"inLedger"`
	Validated       bool   `json:"validated"`
	Meta            struct {
		TransactionResult string      `json:"TransactionResult"`
		Amount            interface{} `json:"delivered_amount"`
	} `json:"meta"`
}

type GetXRPTxIssuedCurrency struct {
	Currency string `json:"currency"`
	Issuer   string `json:"issuer"`
	Value    string `json:"value"`
}

func GetXRPTx(instructions Instructions, basicAuth BasicAuth) ([]byte, uint64, bool) {
	data := GetXRPTxRequestPayload{
		Method: "tx",
		Params: []GetXRPTxRequestParams{
			GetXRPTxRequestParams{
				Transaction: hex.EncodeToString(instructions.TxId),
				Binary:      false,
			},
		},
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return []byte{}, 0, true
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", basicAuth.ChainURL, body)
	if err != nil {
		return []byte{}, 0, true
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, 0, true
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return []byte{}, 0, true
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, true
	}
	var checkErrorResp map[string]CheckXRPErrorResponse
	err = json.Unmarshal(respBody, &checkErrorResp)
	if err != nil {
		return []byte{}, 0, true
	}
	respErrString := checkErrorResp["result"].Error
	if respErrString != "" {
		if respErrString == "amendmentBlocked" ||
			respErrString == "failedToForward" ||
			respErrString == "invalid_API_version" ||
			respErrString == "noClosed" ||
			respErrString == "noCurrent" ||
			respErrString == "noNetwork" ||
			respErrString == "tooBusy" {
			return []byte{}, 0, true
		} else {
			return []byte{}, 0, false
		}
	}
	var jsonResp map[string]GetXRPTxResponse
	err = json.Unmarshal(respBody, &jsonResp)
	if err != nil {
		return []byte{}, 0, false
	}
	if jsonResp["result"].TransactionType != "Payment" || !jsonResp["result"].Validated || jsonResp["result"].Meta.TransactionResult != "tesSUCCESS" {
		return []byte{}, 0, false
	}
	inLedger := uint64(jsonResp["result"].InLedger)
	if inLedger == 0 || inLedger >= instructions.AvailableLedger || !jsonResp["result"].Validated {
		return []byte{}, 0, false
	}
	var currency string
	var amount uint64
	if stringAmount, ok := jsonResp["result"].Meta.Amount.(string); ok {
		amount, err = strconv.ParseUint(stringAmount, 10, 64)
		if err != nil {
			return []byte{}, 0, false
		}
		currency = "xrp"
	} else {
		amountStruct, err := json.Marshal(jsonResp["result"].Meta.Amount)
		if err != nil {
			return []byte{}, 0, false
		}
		var issuedCurrencyResp GetXRPTxIssuedCurrency
		err = json.Unmarshal(amountStruct, &issuedCurrencyResp)
		if err != nil {
			return []byte{}, 0, false
		}
		floatAmount, err := strconv.ParseFloat(issuedCurrencyResp.Value, 64)
		if err != nil {
			return []byte{}, 0, false
		}
		amount = uint64(floatAmount * math.Pow(10, 15))
		currency = issuedCurrencyResp.Currency + issuedCurrencyResp.Issuer
	}
	salt := crypto.Keccak256([]byte("FlareStateConnector_PAYMENTHASH"))
	destinationHash := crypto.Keccak256([]byte(jsonResp["result"].Destination))
	destinationTagHash := crypto.Keccak256(common.LeftPadBytes(common.FromHex(hexutil.EncodeUint64(uint64(jsonResp["result"].DestinationTag))), 32))
	destinationHash = crypto.Keccak256(destinationHash, destinationTagHash)
	amountHash := crypto.Keccak256(common.LeftPadBytes(common.FromHex(hexutil.EncodeUint64(uint64(amount))), 32))
	dataHash := crypto.Keccak256([]byte(currency))
	return crypto.Keccak256(salt, destinationHash, dataHash, amountHash), inLedger, false
}

func ProvePaymentXRP(instructions Instructions, basicAuth BasicAuth) (bool, bool) {
	paymentHash, inLedger, err := GetXRPTx(instructions, basicAuth)
	if err {
		return false, true
	}
	if len(paymentHash) > 0 && bytes.Equal(paymentHash, instructions.PaymentHash) && inLedger == instructions.Ledger {
		return true, false
	}
	return false, false
}

// =======================================================
// ALGO
// =======================================================

func ProvePaymentALGO(instructions Instructions, basicAuth BasicAuth) (bool, bool) {
	return false, false
}

// =======================================================
// Common
// =======================================================

func ReadChain(instructions Instructions, basicAuth BasicAuth) (bool, bool) {
	switch instructions.ChainId {
	case 0, 1, 2:
		return ProvePaymentPOW(instructions, basicAuth)
	case 3:
		return ProvePaymentXRP(instructions, basicAuth)
	case 4:
		return ProvePaymentALGO(instructions, basicAuth)
	}
	return false, false
}

func GetAPIs(chainId uint64) string {
	switch chainId {
	case 0:
		return os.Getenv("BTC_APIs")
	case 1:
		return os.Getenv("LTC_APIs")
	case 2:
		return os.Getenv("DOGE_APIs")
	case 3:
		return os.Getenv("XRP_APIs")
	case 4:
		return os.Getenv("ALGO_APIs")
	default:
		return ""
	}
}

type BasicAuth struct {
	ChainURL string
	Username string
	Password string
}

func GetBasicAuth(chainId uint64, chainURL string) BasicAuth {
	var username, password string
	chainURLhash := sha256.Sum256([]byte(chainURL))
	chainURLchecksum := hex.EncodeToString(chainURLhash[0:4])
	switch chainId {
	case 0:
		username = os.Getenv("BTC_U_" + chainURLchecksum)
		password = os.Getenv("BTC_P_" + chainURLchecksum)
	case 1:
		username = os.Getenv("LTC_U_" + chainURLchecksum)
		password = os.Getenv("LTC_P_" + chainURLchecksum)
	case 2:
		username = os.Getenv("DOGE_U_" + chainURLchecksum)
		password = os.Getenv("DOGE_P_" + chainURLchecksum)
	case 3:
	case 4:
	default:
	}
	return BasicAuth{
		ChainURL: chainURL,
		Username: username,
		Password: password,
	}
}

func ReadChainWithRetries(blockTime *big.Int, instructions Instructions) bool {
	switch blockTime {
	default:
		chainURLs := GetAPIs(instructions.ChainId)
		if chainURLs == "" {
			return false
		}
		for i := 0; i < apiRetries; i++ {
			for _, chainURL := range strings.Split(chainURLs, ",") {
				if chainURL == "" {
					continue
				}
				basicAuth := GetBasicAuth(instructions.ChainId, chainURL)
				verified, err := ReadChain(instructions, basicAuth)
				if !verified && err {
					continue
				}
				return verified
			}
			time.Sleep(apiRetryDelay)
		}
		return false
	}
}

func GetVerificationPaths(checkRet []byte) (string, string) {
	prefix := "cache/"
	acceptedPrefix := "ACCEPTED"
	rejectedPrefix := "REJECTED"
	verificationHash := hex.EncodeToString(checkRet)
	suffix := "_" + verificationHash
	return prefix + acceptedPrefix + suffix, prefix + rejectedPrefix + suffix
}

type Instructions struct {
	InitialCommit   bool
	ChainId         uint64
	Ledger          uint64
	Utxo            uint16
	AvailableLedger uint64
	TxId            []byte
	PaymentHash     []byte
}

func ParseInstructions(checkRet []byte, availableLedger uint64) Instructions {
	return Instructions{
		InitialCommit:   binary.BigEndian.Uint64(checkRet[0:8]) == 1,
		ChainId:         binary.BigEndian.Uint64(checkRet[8:16]),
		Ledger:          binary.BigEndian.Uint64(checkRet[16:24]),
		Utxo:            binary.BigEndian.Uint16(checkRet[30:32]),
		AvailableLedger: availableLedger,
		TxId:            checkRet[32:64],
		PaymentHash:     checkRet[64:96],
	}
}

// Verify proof against underlying chain
func StateConnectorCall(blockTime *big.Int, checkRet []byte, availableLedger uint64) bool {
	if len(checkRet) != 96 {
		return false
	}
	instructions := ParseInstructions(checkRet, availableLedger)
	if instructions.InitialCommit {
		go func() {
			acceptedPath, rejectedPath := GetVerificationPaths(checkRet[8:])
			_, errACCEPTED := os.Stat(acceptedPath)
			if errACCEPTED != nil {
				if ReadChainWithRetries(blockTime, instructions) {
					verificationHashStore, err := os.Create(acceptedPath)
					verificationHashStore.Close()
					if err != nil {
						// Permissions problem
						panic(err)
					}
				} else {
					verificationHashStore, err := os.Create(rejectedPath)
					verificationHashStore.Close()
					if err != nil {
						// Permissions problem
						panic(err)
					}
				}
			}
		}()
		return true
	} else {
		acceptedPath, rejectedPath := GetVerificationPaths(checkRet[8:])
		_, errACCEPTED := os.Stat(acceptedPath)
		_, errREJECTED := os.Stat(rejectedPath)
		if errACCEPTED != nil && errREJECTED != nil {
			for i := 0; i < 2*apiRetries; i++ {
				_, errACCEPTED = os.Stat(acceptedPath)
				_, errREJECTED = os.Stat(rejectedPath)
				if errACCEPTED == nil || errREJECTED == nil {
					break
				}
				time.Sleep(apiRetryDelay)
			}
		}
		return errACCEPTED == nil
	}
}
