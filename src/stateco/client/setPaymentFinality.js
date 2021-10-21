// (c) 2021, Flare Networks Limited. All rights reserved.
// Please see the file LICENSE for licensing terms.

'use strict';
process.env.NODE_ENV = 'production';
const Web3 = require('web3');
const web3 = new Web3();
const Tx = require('ethereumjs-tx').Transaction;
const Common = require('ethereumjs-common').default;
const fs = require('fs');
const fetch = require('node-fetch');

var config,
	customCommon,
	stateConnector,
	api,
	username,
	password;

async function postData(url = '', username = '', password = '', data = {}) {
	const response = await fetch(url, {
		method: 'POST',
		headers: new fetch.Headers({
			'Authorization': 'Basic ' + Buffer.from(username + ':' + password).toString('base64'),
			'Content-Type': "application/json"
		}),
		credentials: 'include',
		body: JSON.stringify(data)
	}).catch(processFailure);
	return response.json();
}

// ===============================================================
// Chain Common Functions
// ===============================================================

async function run(chainId) {
	if (chainId >= 0 && chainId < 3) {
		const method = 'getrawtransaction';
		const params = [txId, true];
		postData(api, username, password, { method: method, params: params })
			.then(tx => {
				const method = 'getblockheader';
				const params = [tx.result.blockhash];
				postData(api, username, password, { method: method, params: params })
					.then(block => {
						const leafPromise = new Promise((resolve, reject) => {
							const amount = Math.floor(parseFloat(tx.result.vout[voutN].value).toFixed(8)*Math.pow(10,8));
							console.log('\nchainId: \t\t', chainId, '\n',
								'ledger: \t\t', block.result.height, '\n',
								'txId: \t\t\t', tx.result.txid, '\n',
								'destination: \t\t', tx.result.vout[voutN].scriptPubKey.addresses[0], '\n',
								'amount: \t\t', amount, '\n');
							const utxo = parseInt(voutN);
							const salt = web3.utils.soliditySha3("FlareStateConnector_PAYMENTHASH");
							const destinationHash = web3.utils.soliditySha3(tx.result.vout[voutN].scriptPubKey.addresses[0]);
							const amountHash = web3.utils.soliditySha3(amount);
							const currencyHash = web3.utils.soliditySha3(chainId);;
							const paymentHash = web3.utils.soliditySha3(salt, destinationHash, currencyHash, amountHash);
							const leaf = {
								"chainId": web3.utils.numberToHex(chainId),
								"txId": '0x' + tx.result.txid,
								"utxo": utxo,
								"ledger": parseInt(block.result.height),
								"destination": destinationHash,
								"amount": amount,
								"currency": currencyHash,
								"paymentHash": paymentHash,
							}
							resolve(leaf);
						})
						leafPromise.then(leaf => {
							stateConnector.methods.getPaymentFinality(
								leaf.chainId,
								leaf.ledger,
								leaf.txId,
								leaf.utxo,
								leaf.destination,
								leaf.currency,
								web3.utils.numberToHex(leaf.amount)).call({
									from: config.testAccount.address,
									gas: config.flare.gas,
									gasPrice: config.flare.gasPrice
								}).catch(() => {
								})
								.then(paymentResult => {
									if (typeof paymentResult != "undefined") {
										if (paymentResult == true) {
											console.log('Payment already proven.');
											setTimeout(() => { return process.exit() }, 2500);
										} else {
											return setPaymentFinality(leaf);
										}
									} else {
										return setPaymentFinality(leaf);
									}
								})
						})
					})
			})
	} else if (chainId == 3) {
		const method = 'tx';
		const params = [{
			'transaction': txId,
			'binary': false
		}];
		postData(api, config.chains.xrp.username, config.chains.xrp.password, { method: method, params: params })
			.then(tx => {
				if (tx.result.TransactionType == 'Payment') {
					const leafPromise = new Promise((resolve, reject) => {
						var destinationTag;
						if (!("DestinationTag" in tx.result)) {
							destinationTag = 0;
						} else {
							destinationTag = parseInt(tx.result.DestinationTag);
						}
						var currency;
						var amount;
						if (typeof tx.result.meta.delivered_amount == "string") {
							currency = "xrp";
							amount = parseInt(tx.result.meta.delivered_amount);
						} else {
							currency = tx.result.meta.delivered_amount.currency + tx.result.meta.delivered_amount.issuer;
							amount = parseFloat(tx.result.meta.delivered_amount.value).toFixed(15)*Math.pow(10,15);
						}
						var utxo = 0;
						console.log('\nchainId: \t\t', chainId, '\n',
							'ledger: \t\t', tx.result.inLedger, '\n',
							'txId: \t\t\t', tx.result.hash, '\n',
							'destination: \t\t', tx.result.Destination, '\n',
							'destinationTag: \t', destinationTag, '\n',
							'amount: \t\t', amount, '\n',
							'currency: \t\t', currency, '\n');
						const salt = web3.utils.soliditySha3("FlareStateConnector_PAYMENTHASH");
						const destinationHash = web3.utils.soliditySha3(web3.utils.soliditySha3(tx.result.Destination), web3.utils.soliditySha3(destinationTag));
						const currencyHash = web3.utils.soliditySha3(currency);
						const amountHash = web3.utils.soliditySha3(amount);
						const paymentHash = web3.utils.soliditySha3(salt, destinationHash, currencyHash, amountHash);
						const leaf = {
							"chainId": chainId,
							"txId": '0x' + tx.result.hash,
							"utxo": utxo,
							"ledger": parseInt(tx.result.inLedger),
							"destination": destinationHash,
							"amount": amount,
							"currency": currencyHash,
							"paymentHash": paymentHash,
						}
						resolve(leaf);
					})

					leafPromise.then(leaf => {
						stateConnector.methods.getPaymentFinality(
							leaf.chainId,
							leaf.ledger,
							leaf.txId,
							leaf.utxo,
							leaf.destination,
							leaf.currency,
							web3.utils.numberToHex(leaf.amount)).call({
								from: config.testAccount.address,
								gas: config.flare.gas,
								gasPrice: config.flare.gasPrice
							}).catch(() => {
							})
							.then(paymentResult => {
								if (typeof paymentResult != "undefined") {
									if (paymentResult == true) {
										console.log('Payment already proven.');
										setTimeout(() => { return process.exit() }, 2500);
									} else {
										return setPaymentFinality(leaf);
									}
								} else {
									return setPaymentFinality(leaf);
								}
							})
					})
				} else {
					console.log('Transaction type not yet supported.');
					setTimeout(() => { return process.exit() }, 2500);
				}
			})
	} else {
		return processFailure('Invalid chainId.');
	}
}

async function setPaymentFinality(leaf) {
	web3.eth.getTransactionCount(config.testAccount.address)
		.then(nonce => {
			return [stateConnector.methods.setPaymentFinality(
				leaf.chainId,
				leaf.ledger,
				leaf.txId,
				leaf.utxo,
				leaf.paymentHash).encodeABI(), nonce];
		})
		.then(txData => {
			var rawTx = {
				nonce: txData[1],
				gasPrice: web3.utils.toHex(parseInt(config.flare.gasPrice)),
				gas: web3.utils.toHex(config.flare.gas),
				to: stateConnector.options.address,
				from: config.testAccount.address,
				data: txData[0]
			};
			var tx = new Tx(rawTx, { common: customCommon });
			var key = Buffer.from(config.testAccount.privateKey, 'hex');
			tx.sign(key);
			var serializedTx = tx.serialize();
			const txHash = web3.utils.sha3(serializedTx);
			console.log('Delivering proof:\t\x1b[33m', txHash, '\x1b[0m');
			web3.eth.getTransaction(txHash)
				.then(txResult => {
					if (txResult == null) {
						web3.eth.sendSignedTransaction('0x' + serializedTx.toString('hex'))
							.on('receipt', receipt => {
								if (receipt.status == false) {
									return processFailure('receipt.status == false');
								} else {
									console.log('Proof delivered:\t \x1b[33m' + receipt.transactionHash + '\x1b[0m');
									setTimeout(() => { return process.exit() }, 2500);
								}
							})
							.on('error', error => {
								return processFailure(error);
							});
					} else {
						return processFailure('Already waiting for this transaction to be delivered.');
					}
				})
		})
}


async function configure(chainName) {
	let rawConfig = fs.readFileSync('config.json');
	config = JSON.parse(rawConfig);
	let chainId = config.chains[chainName].chainId
	if (chainId == 0) {
		api = config.chains.btc.api;
		username = config.chains.btc.username;
		password = config.chains.btc.password;
	} else if (chainId == 1) {
		api = config.chains.ltc.api;
		username = config.chains.ltc.username;
		password = config.chains.ltc.password;
	} else if (chainId == 2) {
		api = config.chains.doge.api;
		username = config.chains.doge.username;
		password = config.chains.doge.password;
	} else if (chainId == 3) {
		api = config.chains.xrp.api;
		username = config.chains.xrp.username;
		password = config.chains.xrp.password;
	}
	web3.setProvider(new web3.providers.HttpProvider(config.flare.url));
	web3.eth.handleRevert = true;
	customCommon = Common.forCustomChain('ropsten',
		{
			name: 'coston',
			networkId: config.flare.chainId,
			chainId: config.flare.chainId,
		},
		'petersburg');
	// Read the compiled contract code
	let source = fs.readFileSync("../../../bin/src/stateco/StateConnector.json");
	let contract = JSON.parse(source);
	// Create Contract proxy class
	stateConnector = new web3.eth.Contract(contract.abi);
	// Smart contract EVM bytecode as hex
	stateConnector.options.data = '0x' + contract.deployedBytecode;
	stateConnector.options.from = config.testAccount.address;
	stateConnector.options.address = config.stateConnectorContract;
	return run(chainId);
}

async function processFailure(error) {
	console.error('error:', error);
	setTimeout(() => { return process.exit() }, 2500);
}

const chainName = process.argv[2];
const txId = process.argv[3];
const voutN = process.argv[4];

if (parseInt(voutN) >= 16) {
	processFailure('Proof-of-work payment index too large, must be lower than 16.');
} else if (parseInt(voutN) < 0) {
	processFailure('UTXO index must be positive.');
}
return configure(chainName);