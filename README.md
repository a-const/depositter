# ETH Staking depositter

## Build

```
make
```

will build .deb file

```
make go-build
```

will create bin file in bin directory

## Usage

### new-contract command

flags:

--url Execution grpc url

--chainid Network chain id

--private Deploy account private key (string without 0x prefix)

--public Deploy account public key (string without 0x prefix)

--deposit-file path to json deposit file

### existing-contract command

flags:

Same as for new-contract but with:

--contract-address address of deposit contract (string without 0x prefix)
