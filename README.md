# Block Specimen Producer (Go Ethereum)

* [Introduction](#bsp_intro)
  * [Resources](#bsp_resources)
* [Architecture](#bsp_arch)
* [Build & Run](#build_run)
  * [Flag Definitions](#flag_definitions)
* [Contributing](CONTRIBUTING.md)

## <span id="bsp_intro">Introduction</span>

Essential to the Covalent Network is the Block Specimen Object (BSO) and the Block Specimen Producer (BSP), a bulk export method that ultimately leads to the generation of a canonical representation of a blockchains historical state. Currently implemented on existing blockchain clients running Geth. It functions currently as an -

1. Blockchain data extractor
1. Blockchain data normalizer

What is ultimately created is a ‘Block Specimen’ (BSO), a universal canonical representation of a blockchains historical state.

There are two further considerations regarding the Block Specimen.

1. The BSP is completely standalone on forks of Geth.
1. The separation of data storage layer from the block execution and distributed consensus functionality leads to better segregation and upgrades of functionality in the blockhain data processing pipeline.

As a result, anyone can run full tracing on the block specimen and accurately recreate the blockchain without access to a blockchain client software.

## <span id="bsp_resources">Resources</span>

Production of Block Specimen Objects (BSOs) forms the core of the network’s data objects specification. These objects are created with the aid of three main pieces of open-source software provided by Covalent for the network’s decentralized stack.

1. [Block Specimen Producer (BSP) - Operator run & deployed](https://github.com/covalenthq/go-ethereum)

1. [BSP Agent - Operator run & deployed](https://github.com/covalenthq/mq-store-agent)

1. [BSP Proof-chain - Covalent operated & pre-deployed](https://github.com/covalenthq/cqt-virtnet)

Please refer to these [instructions](https://docs.google.com/document/d/1BMC9-VXZfpB6mGczSu8ylUXJZ_CIx4ephepDtlruv_Q/edit?usp=sharing) for running the BSP with the mq-store-agent (BSP Agent).

Please refer to this [whitepaper](https://docs.google.com/document/d/1J6RalVVfMSh2kSKNHM3Agb4GngzWVw9e1PqLSVb3-PU/edit#) to understand more about its function.

BSP workshop [deck](https://docs.google.com/presentation/d/1qInReJcMxvVywJ8onoFPoKCwuorJ8LpOn3hwLJIl7bg/edit?usp=sharing) for BSP operators.

## <span id="bsp_why">Raison d'être</span>


The blockchain space has been and will continue to be laser-focused on *write* scalability. That is, actually writing on the blockchain (confirming a transaction) and doing so efficiently. And the projects tackling this issue, be it Layer 1s or Layer 2s, have certainly made operating in this space more accommodating, leading to increased adoption as powerful and scalable applications are being developed.

However, this is only one side of the scalability issue that troubles the space. On the flip side you have the issue of *read scalability*. This is different to *write* as the focus with *read* is on extracting and reading the data on the blockchain, whether that be Ethereum, Avalanche or Solana.

One common method of reading data from Ethereum for example is the JSON RPC Layer. A number of issues present themselves however when doing so.

- **Slow**: One needs to make a series of individual data queries to extract the block and its constituent elements like transactions and receipts.
- **Not Multiversion**: Multiversion concurrency control methods are traditionally employed in databases to ensure point-in-time consistent views if multiple parties are viewing or querying the database. Such methods do not exist in web3.
- **Expensive**: To access historical data at any point in time, you need to run your blockchain clients in a mode known as “full archive nodes” - which requires specialized and expensive hardware to scale.
- **The Purge:** For Ethereum specifically, Vitalik recently outlined an updated roadmap for its development which included a phase titled ‘The Purge’. Once this phase is implemented, clients will no longer store historical data older than a year. Hence, alternatives will be needed to access Ethereums full historical state.

Meanwhile, data mappers and static dashboards are great for examining specific metrics and small tables (so long as the smart contract is decoded) but lack flexibility and scalability. Our belief is that -

1. fast, cheap and accessible read capabilities will lead to more diverse and better-adapted blockchain technologies.

1. should be accessible to all, no matter the skill level.

The Block Specimen is **the solution** to tackle the read scalability issues that currently plague blockchains.

## <span id="bsp_arch">Architecture</span>

![diagram](arch.jpg)

While BSOs are currently being created internally at Covalent for each respective blockchain indexed, the Covalent Network shifts this responsibility to operators (anyone performing a role on the Covalent Network). Any operator on the network will be able to opt in to act as a Block Specimen Producer (BSP).

To ensure that the data within the block-specimens that operators create is reliable and honest, a production proof is created for every BSO produced. These will be published to proofing contract deployed by Covalent. Therefore, BSO proofs can be compared and any deviations in the data either accidentally or malicious will have mismatching proofs.

In sum, it is the responsibility of the BSPs to consume blocks from external blockchains and publish both the BSP along with a production proof to the Covalent virtual chain. BSPs play a pivotal role in the network given that the data in the BSO will feed the entire network with the data needed to answer user queries.Of course, operators who successfully perform this role will be compensated in CQT.

## <span id="build_run">Build & Run</span>

Clone the `covalenthq/go-ethereum` repo and checkout the branch that contains the block specimen patch aka `covalent`

```bash
git clone git@github.com:covalenthq/go-ethereum.git
cd go-ethereum
git checkout covalent
```

Build `geth` from source (install `Go` if you don’t have it) and other geth developer tools from root (if you need all the geth related development tools do a `make all`)

```bash
make geth
```

Start redis (our streaming service) with the following.

```bash
brew services start redis          
Successfully started `redis` (label: homebrew.mxcl.redis)
```

Start redis-cli in a separate terminal so you can see the encoded bsps as they are fed into redis streams.

```bash
redis-cli                          
127.0.0.1:6379>
```

We are now ready to start accepting stream message into redis locally

Now start `geth` from root with the given configuration, here we specify the replication targets (block specimen targets) with redis stream topic key `replication`, running `geth` in full `syncmode`, exposing the http port for the geth apis are optional.

Prior to executing, please replace `<user>` with correct local username within the `--datadir` flag. Everything else remains the same as given below.

```bash
./build/bin/geth \
  --mainnet \
  --port 0 \
  --log.debug \
  --syncmode full \
  --datadir /Users/<user>/Library/Ethereum/bsp/ \
  --replication.targets "redis://localhost:6379/?topic=replication" \
  --replica.result \
  --replica.specimen
```

Expect to see the following logs in approx ~ 10 mins as the node begins to sync and export BSOs

```log
blocks=1 txs=0 mgas=0.000 elapsed=12.947ms    mgasps=0.000 number=3 hash=3d6122..8cf741 age=6y4mo1w  dirty=8.92KiB
INFO [11-18|17:24:35.977|core/block_replica.go:112]        Exporting full block-replica
INFO [11-18|17:24:35.977|core/block_replica.go:36]         Creating block replication event         block number=41042 hash=0x0b8706384cf93820c7f8fe72b5463e756d917c94b0df98f850820593b9422b09
```

The last two lines above show that new block replicas containing the block specimens are being produced and streamed to the redis topic “replication”.

Please note it may take anywhere from 2-10 mins to reach this point depending on the strength of the network and other factors that affect the Ethereum network p2p protocol performance.

After this you can check that redis is stacking up the bsp messages through the redis-cli with the command below (this should give you a bunch of messages from the stream)

```bash
127.0.0.1:6379>  XREAD COUNT 4 STREAMS replication 0-0
```

If it doesn’t - the BSP - producer isn't producing messages! In this case please look at the logs above and see if you have any WARN / DEBUG logs that can be responsible for the inoperation.

### <span id="flag_definitions">Flag definitions</span>

`--mainnet` - lets geth know which network to synchronize with, and pull block specimen object from, this can be `--ropsten`,  `--goerli` , `--mainnet` etc

`--port 0` - will auto-assign a port for geth to talk to other nodes in the network, but this may not work if you are behind a firewall. It would be better to explicitly assign a port and to ensure that port is open to any firewalls.

`--http` - enables the json-rpc api over http

`--log.debug` - enables a detailed log of the processes geth deals with going back and forth between

`--syncmode full` - this flag is used to enable different syncing strategies for geth and a fully sync allows us to execute every block from block 0

`--datadir` - specifies a local datadir path for geth (note we use “BSP” as the directory name with the Ethereum directory), this way we don’t overwrite or touch other previously synced geth libraries across other chains

`--replication.targets` - this flag lets the BSP know where and how to send the BSP messages (this flag will not function without the usage of either one or both of the flags below, if both are selected a full block-replica is exported

`--replica.result` - this flag lets the BSP know if all fields related to the block-result specification need to be exported (if only this flag is selected the exported object is a block-result)

`--replica.specimen` -  this flag lets the BSP know if all fields related to the block-specimen specification need to be exported (if only this flag is selected the exported object is a block-specimen)

If both `--replica-result` & `--replica-specimen` are selected then a `block-replica` is exported containing all the fields for exporting any block fully alongwith its stored state.

## Go Ethereum

Official Golang implementation of the Ethereum protocol.

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://pkg.go.dev/github.com/ethereum/go-ethereum?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethereum/go-ethereum)](https://goreportcard.com/report/github.com/ethereum/go-ethereum)
[![Travis](https://travis-ci.com/ethereum/go-ethereum.svg?branch=master)](https://travis-ci.com/ethereum/go-ethereum)
[![Discord](https://img.shields.io/badge/discord-join%20chat-blue.svg)](https://discord.gg/nthXNEv)

Automated builds are available for stable releases and the unstable master branch. Binary
archives are published at https://geth.ethereum.org/downloads/.

## Building the source

For prerequisites and detailed build instructions please read the [Installation Instructions](https://geth.ethereum.org/docs/install-and-build/installing-geth).

Building `geth` requires both a Go (version 1.14 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run

```shell
make geth
```

or, to build the full suite of utilities:

```shell
make all
```

## Executables

The go-ethereum project comes with several wrappers/executables found in the `cmd`
directory.

|    Command    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| :-----------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
|  **`geth`**   | Our main Ethereum CLI client. It is the entry point into the Ethereum network (main-, test- or private net), capable of running as a full node (default), archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the Ethereum network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `geth --help` and the [CLI page](https://geth.ethereum.org/docs/interface/command-line-options) for command line options.          |
|   `clef`    | Stand-alone signing tool, which can be used as a backend signer for `geth`.  |
|   `devp2p`    | Utilities to interact with nodes on the networking layer, without running a full blockchain. |
|   `abigen`    | Source code generator to convert Ethereum contract definitions into easy to use, compile-time type-safe Go packages. It operates on plain [Ethereum contract ABIs](https://docs.soliditylang.org/en/develop/abi-spec.html) with expanded functionality if the contract bytecode is also available. However, it also accepts Solidity source files, making development much more streamlined. Please see our [Native DApps](https://geth.ethereum.org/docs/dapp/native-bindings) page for details. |
|  `bootnode`   | Stripped down version of our Ethereum client implementation that only takes part in the network node discovery protocol, but does not run any of the higher level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks.                                                                                                                                                                                                                                                                 |
|     `evm`     | Developer utility version of the EVM (Ethereum Virtual Machine) that is capable of running bytecode snippets within a configurable environment and execution mode. Its purpose is to allow isolated, fine-grained debugging of EVM opcodes (e.g. `evm --code 60ff60ff --debug run`).                                                                                                                                                                                                                                                                     |
|   `rlpdump`   | Developer utility tool to convert binary RLP ([Recursive Length Prefix](https://eth.wiki/en/fundamentals/rlp)) dumps (data encoding used by the Ethereum protocol both network as well as consensus wise) to user-friendlier hierarchical representation (e.g. `rlpdump --hex CE0183FFFFFFC4C304050583616263`).                                                                                                                                                                                                                                 |
|   `puppeth`   | a CLI wizard that aids in creating a new Ethereum network.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |

## Running `geth`

Going through all the possible command line flags is out of scope here (please consult our
[CLI Wiki page](https://geth.ethereum.org/docs/interface/command-line-options)),
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `geth` instance.

### Full node on the main Ethereum network

By far the most common scenario is people wanting to simply interact with the Ethereum
network: create accounts; transfer funds; deploy and interact with contracts. For this
particular use-case the user doesn't care about years-old historical data, so we can
fast-sync quickly to the current state of the network. To do so:

```shell
$ geth console
```

This command will:
 * Start `geth` in snap sync mode (default, can be changed with the `--syncmode` flag),
   causing it to download more data in exchange for avoiding processing the entire history
   of the Ethereum network, which is very CPU intensive.
 * Start up `geth`'s built-in interactive [JavaScript console](https://geth.ethereum.org/docs/interface/javascript-console),
   (via the trailing `console` subcommand) through which you can interact using [`web3` methods](https://web3js.readthedocs.io/en/) 
   (note: the `web3` version bundled within `geth` is very old, and not up to date with official docs),
   as well as `geth`'s own [management APIs](https://geth.ethereum.org/docs/rpc/server).
   This tool is optional and if you leave it out you can always attach to an already running
   `geth` instance with `geth attach`.

### A Full node on the Görli test network

Transitioning towards developers, if you'd like to play around with creating Ethereum
contracts, you almost certainly would like to do that without any real money involved until
you get the hang of the entire system. In other words, instead of attaching to the main
network, you want to join the **test** network with your node, which is fully equivalent to
the main network, but with play-Ether only.

```shell
$ geth --goerli console
```

The `console` subcommand has the exact same meaning as above and they are equally
useful on the testnet too. Please, see above for their explanations if you've skipped here.

Specifying the `--goerli` flag, however, will reconfigure your `geth` instance a bit:

 * Instead of connecting the main Ethereum network, the client will connect to the Görli
   test network, which uses different P2P bootnodes, different network IDs and genesis
   states.
 * Instead of using the default data directory (`~/.ethereum` on Linux for example), `geth`
   will nest itself one level deeper into a `goerli` subfolder (`~/.ethereum/goerli` on
   Linux). Note, on OSX and Linux this also means that attaching to a running testnet node
   requires the use of a custom endpoint since `geth attach` will try to attach to a
   production node endpoint by default, e.g.,
   `geth attach <datadir>/goerli/geth.ipc`. Windows users are not affected by
   this.

*Note: Although there are some internal protective measures to prevent transactions from
crossing over between the main network and test network, you should make sure to always
use separate accounts for play-money and real-money. Unless you manually move
accounts, `geth` will by default correctly separate the two networks and will not make any
accounts available between them.*

### Full node on the Rinkeby test network

Go Ethereum also supports connecting to the older proof-of-authority based test network
called [*Rinkeby*](https://www.rinkeby.io) which is operated by members of the community.

```shell
$ geth --rinkeby console
```

### Full node on the Ropsten test network

In addition to Görli and Rinkeby, Geth also supports the ancient Ropsten testnet. The
Ropsten test network is based on the Ethash proof-of-work consensus algorithm. As such,
it has certain extra overhead and is more susceptible to reorganization attacks due to the
network's low difficulty/security.

```shell
$ geth --ropsten console
```

*Note: Older Geth configurations store the Ropsten database in the `testnet` subdirectory.*

### Configuration

As an alternative to passing the numerous flags to the `geth` binary, you can also pass a
configuration file via:

```shell
$ geth --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to
export your existing configuration:

```shell
$ geth --your-favourite-flags dumpconfig
```

*Note: This works only with `geth` v1.6.0 and above.*

#### Docker quick start

One of the quickest ways to get Ethereum up and running on your machine is by using
Docker:

```shell
docker run -d --name ethereum-node -v /Users/alice/ethereum:/root \
           -p 8545:8545 -p 30303:30303 \
           ethereum/client-go
```

This will start `geth` in fast-sync mode with a DB memory allowance of 1GB just as the
above command does.  It will also create a persistent volume in your home directory for
saving your blockchain as well as map the default ports. There is also an `alpine` tag
available for a slim version of the image.

Do not forget `--http.addr 0.0.0.0`, if you want to access RPC from other containers
and/or hosts. By default, `geth` binds to the local interface and RPC endpoints is not
accessible from the outside.

### Programmatically interfacing `geth` nodes

As a developer, sooner rather than later you'll want to start interacting with `geth` and the
Ethereum network via your own programs and not manually through the console. To aid
this, `geth` has built-in support for a JSON-RPC based APIs ([standard APIs](https://eth.wiki/json-rpc/API)
and [`geth` specific APIs](https://geth.ethereum.org/docs/rpc/server)).
These can be exposed via HTTP, WebSockets and IPC (UNIX sockets on UNIX based
platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by `geth`,
whereas the HTTP and WS interfaces need to manually be enabled and only expose a
subset of APIs due to security reasons. These can be turned on/off and configured as
you'd expect.

HTTP based JSON-RPC API options:

  * `--http` Enable the HTTP-RPC server
  * `--http.addr` HTTP-RPC server listening interface (default: `localhost`)
  * `--http.port` HTTP-RPC server listening port (default: `8545`)
  * `--http.api` API's offered over the HTTP-RPC interface (default: `eth,net,web3`)
  * `--http.corsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--ws.addr` WS-RPC server listening interface (default: `localhost`)
  * `--ws.port` WS-RPC server listening port (default: `8546`)
  * `--ws.api` API's offered over the WS-RPC interface (default: `eth,net,web3`)
  * `--ws.origins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: `admin,debug,eth,miner,net,personal,shh,txpool,web3`)
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to
connect via HTTP, WS or IPC to a `geth` node configured with the above flags and you'll
need to speak [JSON-RPC](https://www.jsonrpc.org/specification) on all transports. You
can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based
transport before doing so! Hackers on the internet are actively trying to subvert
Ethereum nodes with exposed APIs! Further, all browser tabs can access locally
running web servers, so malicious web pages could try to subvert locally available
APIs!**

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for
granted in the official networks need to be manually set up.

#### Defining the private genesis state

First, you'll need to create the genesis state of your networks, which all nodes need to be
aware of and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):

```json
{
  "config": {
    "chainId": <arbitrary positive integer>,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0
  },
  "alloc": {},
  "coinbase": "0x0000000000000000000000000000000000000000",
  "difficulty": "0x20000",
  "extraData": "",
  "gasLimit": "0x2fefd8",
  "nonce": "0x0000000000000042",
  "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp": "0x00"
}
```

The above fields should be fine for most purposes, although we'd recommend changing
the `nonce` to some random value so you prevent unknown remote nodes from being able
to connect to you. If you'd like to pre-fund some accounts for easier testing, create
the accounts and populate the `alloc` field with their addresses.

```json
"alloc": {
  "0x0000000000000000000000000000000000000001": {
    "balance": "111111111"
  },
  "0x0000000000000000000000000000000000000002": {
    "balance": "222222222"
  }
}
```

With the genesis state defined in the above JSON file, you'll need to initialize **every**
`geth` node with it prior to starting it up to ensure all blockchain parameters are correctly
set:

```shell
$ geth init path/to/genesis.json
```

#### Creating the rendezvous point

With all nodes that you want to run initialized to the desired genesis state, you'll need to
start a bootstrap node that others can use to find each other in your network and/or over
the internet. The clean way is to configure and run a dedicated bootnode:

```shell
$ bootnode --genkey=boot.key
$ bootnode --nodekey=boot.key
```

With the bootnode online, it will display an [`enode` URL](https://eth.wiki/en/fundamentals/enode-url-format)
that other nodes can use to connect to it and exchange peer information. Make sure to
replace the displayed IP address information (most probably `[::]`) with your externally
accessible IP to get the actual `enode` URL.

*Note: You could also use a full-fledged `geth` node as a bootnode, but it's the less
recommended way.*

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try
`telnet <ip> <port>` to ensure it's indeed reachable), start every subsequent `geth`
node pointed to the bootnode for peer discovery via the `--bootnodes` flag. It will
probably also be desirable to keep the data directory of your private network separated, so
do also specify a custom `--datadir` flag.

```shell
$ geth --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

*Note: Since your network will be completely cut off from the main and test networks, you'll
also need to configure a miner to process transactions and create new blocks for you.*

#### Running a private miner

Mining on the public Ethereum network is a complex task as it's only feasible using GPUs,
requiring an OpenCL or CUDA enabled `ethminer` instance. For information on such a
setup, please consult the [EtherMining subreddit](https://www.reddit.com/r/EtherMining/)
and the [ethminer](https://github.com/ethereum-mining/ethminer) repository.

In a private network setting, however a single CPU miner instance is more than enough for
practical purposes as it can produce a stable stream of blocks at the correct intervals
without needing heavy resources (consider running on a single thread, no need for multiple
ones either). To start a `geth` instance for mining, run it with all your usual flags, extended
by:

```shell
$ geth <usual-flags> --mine --miner.threads=1 --miner.etherbase=0x0000000000000000000000000000000000000000
```

Which will start mining blocks and transactions on a single CPU thread, crediting all
proceedings to the account specified by `--miner.etherbase`. You can further tune the mining
by changing the default gas limit blocks converge to (`--miner.targetgaslimit`) and the price
transactions are accepted at (`--miner.gasprice`).

## Contribution

Thank you for considering to help out with the source code! We welcome contributions
from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-ethereum, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit
more complex changes though, please check up with the core devs first on [our Discord Server](https://discord.gg/invite/nthXNEv)
to ensure those changes are in line with the general philosophy of the project and/or get
some early feedback which can make both your efforts much lighter as well as our review
and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting)
   guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary)
   guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

Please see the [Developers' Guide](https://geth.ethereum.org/docs/developers/devguide)
for more details on configuring your environment, managing project dependencies, and
testing procedures.

## License

The go-ethereum library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The go-ethereum binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
