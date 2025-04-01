const fs = require('fs');
const path = require('path');
const WebSocket = require('ws');
const Peer = require('simple-peer');
const wrtc = require('@roamhq/wrtc');
const axios = require('axios');
const os = require('os');

const wss = new WebSocket.Server({ port: 2502 });

wss.on('connection', ws => {
  console.log('Client connected');

  ws.on('message', message => {
    if (message.toString().startsWith('INIT_')) {
      const runID = message.toString().split('_')[1];
      ws.send('ACK');

      handle(runID).catch(err => console.error('Error in handle:', err));
    }
  });
});

async function handle(runID) {
  const peer = new Peer({ initiator: true, wrtc });

  peer.on('signal', data => {
    console.log('WebRTC signal recv:', data);

    ws.send(JSON.stringify({ type: 'signal', data }));

    console.log('WebRTC signal sent:', data);
  });

  ws.on('message', message => {
    const msg = JSON.parse(message);
    if (msg.type === 'signal') {
      peer.signal(msg.data);
    }
  });

  peer.on('connect', async () => {
    console.log('WebRTC connected, streaming file...');
    const filePath = path.join(__dirname, '..', 'assets', 'sample_video.mp4');

    const transferStartUnix = Date.now();
    const cpuBefore = await getCpuUsagePercentage();
    const ramBefore = await getRamUsageBytes();
    let cpuWhile = 0;
    let ramWhile = 0;

    setTimeout(async () => {
      cpuWhile = await getCpuUsagePercentage();
      ramWhile = await getRamUsageBytes();
    }, 500);

    let chunks = 0;
    let procs = 0;
    let canFinish = false;
    const finish = () => {
      if (procs !== chunks || !canFinish) {
        return;
      }

      peer.send(Buffer.from('__EOF__'));
      console.log('File sent');

      (async () => {
        const cpuAfter = await getCpuUsagePercentage();
        const ramAfter = await getRamUsageBytes();

        await collectMetrics(runID, {
          "TransferStartUnix": transferStartUnix,
          "BytesPayload": fs.statSync(filePath).size,
          "CpuServerPercentBefore": cpuBefore,
          "CpuServerPercentWhile": cpuWhile,
          "CpuServerPercentAfter": cpuAfter,
          "RamServerBytesBefore": ramBefore,
          "RamServerBytesWhile": ramWhile,
          "RamServerBytesAfter": ramAfter
        });
      })();
    };

    const stream = fs.createReadStream(filePath);
    stream.on('data', chunk => {
      chunks++;

      const sendChunk = () => {
        if (peer.bufferSize > 1_000_000) {
          setTimeout(sendChunk, 50);
        } else {
          peer.send(chunk);
          procs++;

          finish();
        }
      };
      sendChunk();
    });
    stream.on('end', () => {
      canFinish = true;
      finish();
    });
  });

  peer.on('close', () => {
    console.log('WebRTC connection closed');
  });

  peer.on('error', err => console.error('Peer error:', err));
}

wss.on('listening', () => console.log('Server started on port 2502'));

async function collectMetrics(runID, data) {
  try {
    await axios.put(`https://thkm25_collect.nauri.io/${runID}/update`, data);
    console.log('[COLLECTOR] Metrics collected!');
  } catch (error) {
    console.error('[COLLECTOR] Error collecting metrics:', error);
  }
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