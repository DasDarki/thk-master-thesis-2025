const fs = require('fs');
const path = require('path');
const WebSocket = require('ws');
const Peer = require('simple-peer');
const wrtc = require('@roamhq/wrtc');

const wss = new WebSocket.Server({ port: 2502 });

wss.on('connection', ws => {
  console.log('Client connected');

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

  peer.on('connect', () => {
    console.log('WebRTC connected, streaming file...');
    const filePath = path.join(__dirname, '..', 'assets', 'sample_video.mp4');

    let chunks = 0;
    let procs = 0;
    let canFinish = false;
    const finish = () => {
      if (procs !== chunks || !canFinish) {
        return;
      }

      peer.send(Buffer.from('__EOF__'));
      console.log('File sent');
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
});

wss.on('listening', () => console.log('Server started on port 2502'));