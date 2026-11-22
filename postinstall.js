#!/usr/bin/env node
const fs = require("fs");
const os = require("os");
const path = require("path");

const appName = "launchbox";

const arch = os.arch(); // "arm64", "x64", "amd64"
const platform = os.platform(); // "darwin", "linux", "win32"

if (platform === "win32") {
  console.warn("⚠️ Windows is not supported natively, use WSL instead. Skipping.");
  process.exit(0);
}

const architecture = arch === "x64" ? "amd64" : arch;
const binaryName = `${appName}-${platform}-${architecture}`;

const source = path.join(__dirname, "bin", binaryName);
const dest = path.join(__dirname, appName);

if (!fs.existsSync(source)) {
  console.warn(`⚠️ Binary not found for ${platform}-${arch}, skipping.`);
  process.exit(0);
}

fs.copyFileSync(source, dest);
fs.chmodSync(dest, 0o755);

console.log(`✅ Installed binary: ${binaryName}`);
