import React from "react";
import { Route, Switch, useLocation } from "wouter";
import { CosmWasmClient } from "@cosmjs/cosmwasm-stargate";

// Constants for configuration
const CONTRACT_ADDRESS =
  "bbn1789sj5p55hrjk6zrz8nzplhde4ae8mht5dnhwsdc5zn9kjw3y4hqndrsrr";
const RPC_ENDPOINT = "https://rpc.sat-bbn-testnet1.satlayer.net";
const OPERATORS = [
  "bbn1lkavyt5gqtv4qu8cufwer5rs4uq2a28emvf24t",
  "bbn1d9878dze7npzf7t3vxh8f5y2munj7a8xuy50m8",
];

const CHAINS = [
  {
    title: "Babylon",
    chainId: "sat-bbn-testnet1",
    evmChainId: 100,
    currency: "BBN",
    iconPath: "/babylon.webp",
    rpc: "https://rpc.sat-bbn-testnet1.satlayer.net",
    rest: "https://lcd1.sat-bbn-testnet1.satlayer.net",
    bech32Config: {
      bech32PrefixAccAddr: "bbn",
      bech32PrefixAccPub: "bbnpub",
      bech32PrefixValAddr: "bbnvaloper",
      bech32PrefixValPub: "bbnvaloperpub",
      bech32PrefixConsAddr: "bbnvalcons",
      bech32PrefixConsPub: "bbnvalconspub",
    },
  },
  {
    title: "Ethereum Mainnet",
    chainId: "ethereum-1",
    evmChainId: 1,
    currency: "ETH",
    iconPath: "/ethereum.webp",
    rpc: "https://eth.llamarpc.com",
  },
  {
    title: "Base",
    chainId: "base-1",
    evmChainId: 8453,
    currency: "ETH",
    iconPath: "/base.webp",
    rpc: "https://base.llamarpc.com",
  },
  {
    title: "Polygon",
    chainId: "polygon-1",
    evmChainId: 137,
    currency: "MATIC",
    iconPath: "/polygon.webp",
    rpc: "https://polygon.llamarpc.com",
  },
];

function App() {
  const [leaderboard, setLeaderboard] = React.useState([]);
  const [client, setClient] = React.useState(null);
  const [error, setError] = React.useState(null);
  const [isInitializing, setIsInitializing] = React.useState(true);
  const [selectedChain, setSelectedChain] = React.useState(CHAINS[0]);

  React.useEffect(() => {
    async function initClient() {
      try {
        setIsInitializing(true);
        setError(null);
        const cosmClient = await CosmWasmClient.connect(RPC_ENDPOINT);
        setClient(cosmClient);
      } catch (error) {
        console.error("Failed to connect to CosmWasm client:", error);
        setError("Failed to connect to the network. Please try again later.");
      } finally {
        setIsInitializing(false);
      }
    }
    initClient();

    return () => {
      if (client) {
        client.disconnect?.();
      }
    };
  }, []);

  React.useEffect(() => {
    if (!client || isInitializing) return;

    async function fetchOperatorData() {
      try {
        setError(null);
        const blockHeight = await client.getHeight();

        const operatorDataPromises = OPERATORS.map(async (operator) => {
          try {
            const scoreResult = await client.queryContractSmart(
              CONTRACT_ADDRESS,
              {
                get_operator_score: { operator },
              }
            );

            const maxScoreResult = await client.queryContractSmart(
              CONTRACT_ADDRESS,
              {
                get_operator_max_score: { operator },
              }
            );

            const score = parseInt(scoreResult);
            const maxScore = parseInt(maxScoreResult);
            const percentage =
              maxScore > 0 ? ((score / maxScore) * 100).toFixed(2) : 0;

            const startTime = Date.now();
            await client.getHeight();
            const latency = Date.now() - startTime;

            return {
              operator,
              score,
              maxScore,
              percentage: parseFloat(percentage),
              latency,
              height: blockHeight,
            };
          } catch (error) {
            console.error(
              `Error fetching data for operator ${operator}:`,
              error
            );
            return {
              operator,
              score: 0,
              maxScore: 0,
              percentage: 0,
              latency: 0,
              height: blockHeight,
              error: true,
            };
          }
        });

        const operatorData = await Promise.all(operatorDataPromises);
        const formattedData = operatorData.map((data) => [
          data.operator,
          data.score,
          data.maxScore,
          data.percentage,
          data.latency,
          data.height,
        ]);

        setLeaderboard(formattedData.sort((a, b) => b[3] - a[3]));
      } catch (error) {
        console.error("Error fetching operator data:", error);
        setError("Failed to fetch operator data. Please try again later.");
      }
    }

    fetchOperatorData();
    const interval = setInterval(fetchOperatorData, 3000);
    return () => clearInterval(interval);
  }, [client, isInitializing]);

  return (
    <>
      <Switch>
        <Route
          path="/"
          component={() => <Homepage onChainSelect={setSelectedChain} />}
        />
        <Route path="/leaderboard">
          <div className="flex items-center justify-center h-screen flex-col mt-[-100px]">
            {error && (
              <div className="w-full max-w-[1024px] px-4 lg:px-0">
                <div className="mt-4 p-4 bg-red-100 text-red-700 rounded-md">
                  {error}
                </div>
              </div>
            )}
            <Leaderboard
              leaderboard={leaderboard}
              isLoading={isInitializing || leaderboard.length === 0}
              selectedChain={selectedChain}
            />
          </div>
        </Route>
      </Switch>
    </>
  );
}

function Homepage({ onChainSelect }) {
  const [, setLocation] = useLocation();

  const handleChainSelect = (chain) => {
    onChainSelect(chain);
    setLocation("/leaderboard");
  };

  return (
    <div className="flex flex-col lg:grid lg:grid-cols-2 min-h-screen">
      <div className="bg-[#FFB800] flex flex-col items-center justify-center p-8 lg:p-0">
        <div className="py-12 lg:py-0">
          <div className="flex flex-col items-center justify-center gap-2">
            <img className="w-[40%]" src="satlayer.svg" alt="" />
            <h1 className="font-bold text-4xl text-[#000000]">satRPC</h1>
            <div className="w-full lg:w-[60%] text-center flex flex-col gap-2">
              <p>
                SatRPC is an BVS that enables the creation and coordination of a
                decentralized, secure, and reliable RPC network for any chain.
              </p>
              <p>
                SatRPC operators work together to ensure the integrity of the
                network, and operator reputation scores ensure users can access
                the most trustworthy nodes.
              </p>
            </div>
          </div>
        </div>
      </div>
      <div className="w-full p-6 h-full">
        <div className="mb-6">
          <h2 className="font-bold text-lg -mb-1">Networks</h2>
          <span className="opacity-40">Select a network</span>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-1 gap-4">
          {CHAINS.map((chain) => (
            <div
              key={chain.chainId}
              className="p-4 flex flex-col border border-gray-100 bg-white items-center gap-2 rounded-lg"
            >
              <img
                className="h-6 rounded-full"
                src={chain.iconPath}
                alt={chain.title}
              />
              <h2 className="text-xl font-bold text-center">{chain.title}</h2>
              <span>
                {chain.evmChainId}{" "}
                <span className="opacity-20 font-mono">/</span> {chain.currency}
              </span>
              <button
                className="border-black border hover:text-white text-black text-center py-2 w-full sm:w-[50%] rounded-full hover:bg-black"
                onClick={() => handleChainSelect(chain)}
              >
                Select
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function Leaderboard({ leaderboard, isLoading, selectedChain }) {
  const [addingWallets, setAddingWallets] = React.useState({});
  const [walletError, setWalletError] = React.useState(null);
  const [walletStatus, setWalletStatus] = React.useState(null);
  const [, setLocation] = useLocation();

  const handleAddToWallet = async (operator, operatorRpc) => {
    if (!window.keplr) {
      setWalletError(
        "Keplr wallet is not installed. Please install Keplr wallet to continue."
      );
      return;
    }

    try {
      setAddingWallets((prev) => ({ ...prev, [operator]: true }));
      setWalletError(null);
      setWalletStatus("Adding network...");

      if (selectedChain.chainId === "sat-bbn-testnet1") {
        const chainInfo = {
          chainId: selectedChain.chainId,
          chainName: "SatLayer Babylon Testnet",
          rpc: selectedChain.rpc,
          rest: selectedChain.rest,
          bip44: {
            coinType: 118,
          },
          bech32Config: selectedChain.bech32Config,
          currencies: [
            {
              coinDenom: "BBN",
              coinMinimalDenom: "ubbn",
              coinDecimals: 6,
            },
          ],
          feeCurrencies: [
            {
              coinDenom: "BBN",
              coinMinimalDenom: "ubbn",
              coinDecimals: 6,
              gasPriceStep: {
                low: 0.01,
                average: 0.025,
                high: 0.04,
              },
            },
          ],
          stakeCurrency: {
            coinDenom: "BBN",
            coinMinimalDenom: "ubbn",
            coinDecimals: 6,
          },
          features: ["ibc-transfer", "ibc-go"],
        };

        await window.keplr.experimentalSuggestChain(chainInfo);
        await window.keplr.enable(chainInfo.chainId);
        const key = await window.keplr.getKey(chainInfo.chainId);
        setWalletStatus(
          `Connected to Babylon Network (${shortenAddress(key.bech32Address)})!`
        );
      } else {
        // For EVM chains, just show the RPC info for now
        setWalletStatus(
          `RPC URL for ${selectedChain.title}: ${selectedChain.rpc}`
        );
      }

      setTimeout(() => setWalletStatus(null), 3000);
    } catch (error) {
      console.error("Failed to add to Keplr:", error);
      setWalletError(
        error.message || "Failed to add network to Keplr. Please try again."
      );
    } finally {
      setAddingWallets((prev) => ({ ...prev, [operator]: false }));
    }
  };

  return (
    <div className="w-full max-w-[1024px] px-4 lg:px-0">
      <div className="mb-4">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h2 className="font-bold text-xl sm:text-2xl flex items-center gap-2">
              <span>
                <img className="h-8 sm:h-10" src="satlayer.svg" alt="" />
              </span>
              <span>satRPC</span>
            </h2>
            <p className="text-sm sm:text-base opacity-80 opacity-40">
              Powered by Babylon and SatLayer Stack
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setLocation("/")}
              className="text-sm border border-black/10 rounded-full px-3 py-1 hover:bg-black hover:text-white transition-colors whitespace-nowrap"
            >
              ← Change Network
            </button>
            <div className="flex items-center gap-2">
              <img
                className="h-5 sm:h-6 rounded-full"
                src={selectedChain.iconPath}
                alt={selectedChain.title}
              />
              <span className="font-bold text-sm sm:text-base">
                {selectedChain.title}
              </span>
            </div>
          </div>
        </div>
      </div>

      {walletError && (
        <div className="bg-red-100 text-red-700 p-3 sm:p-4 mb-4 rounded-md text-sm sm:text-base">
          {walletError}
        </div>
      )}
      {walletStatus && !walletError && (
        <div className="bg-blue-100 text-blue-700 p-3 sm:p-4 mb-4 rounded-md text-sm sm:text-base">
          {walletStatus}
        </div>
      )}
      <div className="w-full bg-white border rounded-lg overflow-hidden">
        <table className="w-full">
          <thead>
            <tr>
              <th className="text-left pl-2 bg-gray-100/50 py-2 text-sm sm:text-base">
                RPC
              </th>
              <th className="hidden sm:table-cell text-left bg-gray-100/50 py-2">
                Address
              </th>
              <th className="hidden sm:table-cell text-left bg-gray-100/50 py-2">
                Latency
              </th>
              <th className="hidden sm:table-cell text-left bg-gray-100/50 py-2">
                Height
              </th>
              <th className="text-left bg-gray-100/50 py-2 text-sm sm:text-base">
                Score
              </th>
              <th className="text-left bg-gray-100/50"></th>
            </tr>
          </thead>
          <tbody>
            {!isLoading &&
              leaderboard.map((l, i) => (
                <tr key={l[0]}>
                  <td className="py-2 border-b border-black/10 pl-2 text-sm sm:text-base">
                    {l[0].slice(0, 6)}.rpc.node
                  </td>
                  <td className="hidden sm:table-cell py-2 border-b border-black/10 font-mono text-sm">
                    {shortenAddress(l[0])}
                  </td>
                  <td className="hidden sm:table-cell py-2 border-b border-black/10 font-mono text-sm">
                    {l[4]?.toFixed(2)} ms
                  </td>
                  <td className="hidden sm:table-cell py-2 border-b border-black/10 font-mono text-sm">
                    {l[5]?.toLocaleString()}
                  </td>
                  <td className="py-2 border-b border-black/10 text-left">
                    <span className="block flex gap-1 items-center">
                      {l[3] > 90 ? (
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          viewBox="0 0 20 20"
                          fill="green"
                          className="w-4 h-4 sm:w-5 sm:h-5"
                        >
                          <path
                            fillRule="evenodd"
                            d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z"
                            clipRule="evenodd"
                          />
                        </svg>
                      ) : (
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          viewBox="0 0 20 20"
                          fill="red"
                          className="w-4 h-4 sm:w-5 sm:h-5"
                        >
                          <path
                            fillRule="evenodd"
                            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z"
                            clipRule="evenodd"
                          />
                        </svg>
                      )}
                      <span
                        className={`text-sm sm:text-base ${
                          l[3] > 90 ? "text-green-500" : "text-red-500"
                        }`}
                      >
                        {l[3].toFixed(2)}%
                      </span>
                    </span>
                  </td>
                  <td className="border-b border-black/10 text-right pr-2">
                    <button
                      onClick={() =>
                        handleAddToWallet(
                          l[0],
                          "https://rpc.sat-bbn-testnet1.satlayer.net"
                        )
                      }
                      disabled={addingWallets[l[0]]}
                      className={`border-black border text-black text-center rounded-full px-2 py-1 text-sm sm:text-base
                        ${addingWallets[l[0]] ? "opacity-50 cursor-not-allowed" : "hover:text-white hover:bg-black"}`}
                    >
                      {addingWallets[l[0]] ? (
                        <span className="flex items-center gap-2">
                          <svg
                            className="animate-spin h-3 w-3 sm:h-4 sm:w-4 text-blue-500"
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                          >
                            <circle
                              className="opacity-25"
                              cx="12"
                              cy="12"
                              r="10"
                              stroke="currentColor"
                              strokeWidth="4"
                            />
                            <path
                              className="opacity-75"
                              fill="currentColor"
                              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                            />
                          </svg>
                          {walletStatus || "Adding..."}
                        </span>
                      ) : (
                        "Add to Wallet"
                      )}
                    </button>
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
        {isLoading && (
          <div className="bg-gray-100/20 p-4 flex items-center justify-center">
            <svg
              className="animate-spin -ml-1 mr-3 h-4 w-4 sm:h-5 sm:w-5"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            <span className="text-sm sm:text-base">Loading...</span>
          </div>
        )}
      </div>
    </div>
  );
}

function shortenAddress(address) {
  return address ? address.slice(0, 6) + "..." + address.slice(-4) : "";
}

export default App;
