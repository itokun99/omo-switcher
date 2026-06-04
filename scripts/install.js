#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');

const PLATFORM = process.platform;
const ARCH = process.arch;

const PLATFORM_MAP = {
  'darwin': 'darwin',
  'linux': 'linux',
  'win32': 'windows'
};

const ARCH_MAP = {
  'x64': 'amd64',
  'arm64': 'arm64'
};

const REPO = 'itokun99/omo-switch';
const VERSION = 'v2.0.0';

function getBinaryName() {
  const platform = PLATFORM_MAP[PLATFORM];
  const arch = ARCH_MAP[ARCH];
  
  if (!platform || !arch) {
    console.error(`Unsupported platform: ${PLATFORM}-${ARCH}`);
    process.exit(1);
  }
  
  const ext = PLATFORM === 'win32' ? '.exe' : '';
  return `omo-switch-${platform}-${arch}${ext}`;
}

function getDownloadUrl() {
  const binary = getBinaryName();
  return `https://github.com/${REPO}/releases/download/${VERSION}/${binary}`;
}

function getInstallPath() {
  const binDir = path.join(__dirname, '..', 'bin');
  const ext = PLATFORM === 'win32' ? '.exe' : '';
  return path.join(binDir, `omo-switch${ext}`);
}

async function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        download(response.headers.location, dest).then(resolve).catch(reject);
        return;
      }
      if (response.statusCode !== 200) {
        reject(new Error(`Download failed: ${response.statusCode}`));
        return;
      }
      response.pipe(file);
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {});
      reject(err);
    });
  });
}

async function main() {
  const url = getDownloadUrl();
  const dest = getInstallPath();
  
  console.log(`Downloading omo-switch for ${PLATFORM}-${ARCH}...`);
  console.log(`URL: ${url}`);
  
  try {
    await download(url, dest);
    fs.chmodSync(dest, '755');
    console.log('Installation complete!');
  } catch (err) {
    console.error('Installation failed:', err.message);
    console.error('');
    console.error('Please install manually:');
    console.error('  go install github.com/itokun99/omo-switch/cmd/omo-switch@latest');
    // Don't fail the install — allow manual install
    process.exit(0);
  }
}

main();
