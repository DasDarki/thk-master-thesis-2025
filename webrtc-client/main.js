const WebSocket = require('ws');
const Peer = require('simple-peer');
const wrtc = require('@roamhq/wrtc');
const fs = require('fs');
const axios = require('axios');
const cp = require('child_process');
const os = require('os');

const [runID, isLocal] = parseArguments();

const ws = new WebSocket(isLocal ? 'ws://localhost:2502' : 'wss://thkm25_websockets.nauri.io');
let peer;

ws.on('open', () => {
  console.log('Connected to signaling server');
  
  ws.send(`INIT_${runID}`);
});

ws.on('message', (message) => {
  if (message.toString() === 'ACK') {
    handle();
  }
});

async function handle() {
  peer = new Peer({ initiator: false, wrtc });

  let connectionEstablished = 0;

  peer.on('signal', data => {
    console.log('WebRTC signal recv:', data);

    ws.send(JSON.stringify({ type: 'signal', data }));

    console.log('WebRTC signal sent:', data);

    if (connectionEstablished === 0) {
      connectionEstablished = Date.now();
    }
  });

  ws.on('message', message => {
    const msg = JSON.parse(message);
    if (msg.type === 'signal') {
      peer.signal(msg.data);
    }
  });

  const cpuBefore = await getCpuUsagePercentage();
  const ramBefore = await getRamUsageBytes();
  let cpuWhile = 0;
  let ramWhile = 0;
  const [lost, recv] = await getPacketStats();

  let first = true;
  const fileStream = fs.createWriteStream(`output${runID}.mp4`);
  peer.on('data', chunk => {
    if (chunk.toString() === '__EOF__') {
      console.log('File fully received');
      fileStream.end(() => console.log('Closed'));
      peer.destroy();

      (async () => {
        const cpuAfter = await getCpuUsagePercentage();
        const ramAfter = await getRamUsageBytes();
        const [lostAfter, recvAfter] = await getPacketStats();

        await collectMetrics({
          "TransferStartUnix": Date.now(),
          "ConnectionDuration": Date.now() - connectionEstablished,
          "CpuClientPercentBefore": cpuBefore,
          "CpuClientPercentWhile": cpuWhile,
          "CpuClientPercentAfter": cpuAfter,
          "RamClientBytesBefore": ramBefore,
          "RamClientBytesWhile": ramWhile,
          "RamClientBytesAfter": ramAfter,
          "LostPackets": lostAfter - lost,
          "BytesSentTotal": recvAfter - recv,
        });
      })();
      return;
    }
  
    console.log('Received chunk...');
    fileStream.write(Buffer.from(chunk));

    if (first) {
      first = false;
      (async () => {
        cpuWhile = await getCpuUsagePercentage();
        ramWhile = await getRamUsageBytes();
      })();
    }
  });
  peer.on('close', () => {
    console.log('WebRTC connection closed');
    ws.close();
  });

  peer.on('error', err => console.error('Peer error:', err));
}

async function collectMetrics(data) {
  try {
    console.log(data);
    await axios.put(`https://thkm25_collect.nauri.io/${runID}/update`, data, {
      headers: {
        'X-API-KEY': 'thk_masterthesis_2025_hwtwswrtc'
      }
    });
    console.log('[COLLECTOR] Metrics collected!');
  } catch (error) {
    console.error('[COLLECTOR] Error collecting metrics:', error);
  }
}

function parseArguments() {
  const args = process.argv;
  const runIDArg = args.find(arg => arg.startsWith('-r'));
  const runID = runIDArg ? runIDArg.substring(2) : null;
  const isLocal = args.includes('-l');

  return [runID, isLocal];
}

function getCpuUsagePercentage() {
  return new Promise((resolve) => {
    const start = cpuInfo();

    setTimeout(() => {
      const end = cpuInfo();

      const idleDiff = end.idle - start.idle;
      const totalDiff = end.total - start.total;

      const usage = 100 - Math.round((idleDiff / totalDiff) * 100);
      resolve(usage);
    }, 100);
  });
}

function cpuInfo() {
  const cpus = os.cpus();

  let idle = 0;
  let total = 0;

  for (const cpu of cpus) {
    for (const type in cpu.times) {
      total += cpu.times[type];
    }
    idle += cpu.times.idle;
  }

  return { idle, total };
}

async function getRamUsageBytes() {
  const total = os.totalmem();
  const free = os.freemem();
  return total - free;
}

// Returns netstat -e output for lost packets, and bytes received (sent)
async function getPacketStats() {
  const netstatOutput = cp.execSync('netstat -e').toString().split('\n');
  let lost = 0;
  let recv = 0;

  for (const line of netstatOutput) {
    if (line.startsWith("Bytes")) {
      const parts = line.trim().split(/\s+/);
      if (parts.length > 0) {
        recv = parseInt(parts[1], 10);
      }
    } else if (line.startsWith("Verworfen") || line.startsWith("Lost") || line.startsWith("Fehler") || line.startsWith("Errors")) {
      const parts = line.trim().split(/\s+/);
      if (parts.length > 0) {
        lost += parseInt(parts[1], 10);
      }
    }
  }

  return [lost, recv];
}