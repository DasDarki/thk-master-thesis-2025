const WebSocket = require('ws');
const Peer = require('simple-peer');
const wrtc = require('@roamhq/wrtc');
const fs = require('fs');

const ws = new WebSocket('ws://localhost:2502');
let peer;

ws.on('open', () => {
  console.log('Connected to signaling server');

  peer = new Peer({ initiator: false, wrtc });

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

  const fileStream = fs.createWriteStream('output.mp4');
  peer.on('data', chunk => {
    if (chunk.toString() === '__EOF__') {
      console.log('File fully received');
      fileStream.end(() => console.log('Closed'));
      peer.destroy();
      return;
    }
  
    console.log('Received chunk...');
    fileStream.write(Buffer.from(chunk));
  });
  peer.on('close', () => {
    console.log('WebRTC connection closed');
    ws.close();
  });

  peer.on('error', err => console.error('Peer error:', err));
});
