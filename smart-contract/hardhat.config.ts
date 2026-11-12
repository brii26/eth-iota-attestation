import { defineConfig } from "hardhat/config";

// Hardhat compiles the contract. Tests, deployment, and bindings are written in Go.
export default defineConfig({
  solidity: {
    version: "0.8.28",
    settings: {
      evmVersion: "cancun",
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
});
