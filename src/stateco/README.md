# State-Connector System: Proving the state of any underlying chain for all smart contracts on Flare

The state connector system is a competitive approach for proving the state of an underlying chain to a smart contract, and it has the following advantages:

1. **Transaction validity references back to an underlying chain's genesis block**: Other approaches like the SPV proof do not check the validity of a transaction.

2. **Safety only depends on an underlying chain's validators**: There is no trusted third-party service that has its own set of economic incentives and risks. Trust is minimized by leveraging the guarantee that safety can only be lost in the state connector if an underlying chain's validators encounter a Byzantine fault.

3. **No cooperation needed from an underlying chain's validators**: Validators from an underlying chain do not need to modify their chain's codebase to permit Flare to interpret their network. An underlying chain's validators do not even need to be aware that Flare exists in order for the state connector system to operate.

4. **Can read the state of any blockchain**: The state connector can operate on any possible Sybil-resistance technique of an underlying chain. For example: proof-of-work, proof-of-stake and even federated byzantine agreement where there is not global agreement on the set of validators in control of a network.

5. **No encoding of the current validators in control of an underlying chain to a smart contract on Flare**: This requirement of other state-relay approaches such as the SPV proof leads to the hazardous scenario where the enforcement of bad behavior in relaying state needs to be conducted by the same set of operators that have performed the bad behavior.

6. **Constant-sized proofs**: both the data availability proof and the payment proof are constant-sized, independent of the number of other payments in the data availability period being considered.

7. **Every Flare validator independently verifies an underlying chain's state**: If your own Flare validator observes the canonical state of an underlying chain, then you will not lose safety against that chain.

# Verify an Underlying Chain Payment on Flare

## Proving a Payment

Once the first data availability proof has been finalised, you can then submit a payment proof regarding the XRP transaction: [FFB44382D074CB37B63AC9D3EB2D829C1D1FE4D54DC1A0BCC1D23BAE18D53272](https://livenet.xrpl.org/transactions/FFB44382D074CB37B63AC9D3EB2D829C1D1FE4D54DC1A0BCC1D23BAE18D53272). Run the following command in a separate terminal window:

```
node setPaymentFinality xrp FFB44382D074CB37B63AC9D3EB2D829C1D1FE4D54DC1A0BCC1D23BAE18D53272
```

The following command proves a Litecoin payment in the first-position of the UTXO output for this transaction: [0956165f77106ad62d42a3236db3e47178adfa7a80cc1fad43b894fa4ed0c581](https://live.blockcypher.com/ltc/tx/0956165f77106ad62d42a3236db3e47178adfa7a80cc1fad43b894fa4ed0c581/)

```
node setPaymentFinality ltc 0956165f77106ad62d42a3236db3e47178adfa7a80cc1fad43b894fa4ed0c581 0
```

## Custom-currency Proofs (e.g. Issued Currencies, ERC20s, etc.)

The proving/disproving of a custom-currency payment is also supported. The state connector supports any issued currency and differentiates them by appending their currency code to its issuer's address, e.g. USDrL7jDKUNmxBG24QsqA6fDUwFwjndgMojje. For example, the following command proves that a [payment of USD](https://livenet.xrpl.org/transactions/8B3FB7F0B5BDAB705FDB152EBA20BF47159898D76812DA80BD367D99206B5C59) issued on the XRPL occurred:

```
node setPaymentFinality xrp 8B3FB7F0B5BDAB705FDB152EBA20BF47159898D76812DA80BD367D99206B5C59
```

This example proves a [BTC issued-currency payment](https://livenet.xrpl.org/transactions/67B3F2CAF2905BC67FEB5417C1C3F9AA941DF8984F1F49EC48D4DCADFAC94418) on the XRPL:

```
node setPaymentFinality xrp 67B3F2CAF2905BC67FEB5417C1C3F9AA941DF8984F1F49EC48D4DCADFAC94418
```

## Two-stage Payment Proof Mechanism

The above commands must be run twice with a `30` second gap in between command runs in order to complete the proving/disproving of a payment. The purpose of this is that it removes the underlying-chain API-call delay from the synchronous EVM execution and instead puts the API-call delay burden on the user proving a payment.