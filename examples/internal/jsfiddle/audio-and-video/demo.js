/* eslint-env browser */

const pc = new RTCPeerConnection({
  iceServers: [{
    urls: 'stun:stun.l.google.com:19302'
  }]
})
const log = msg => {
  document.getElementById('logs').innerHTML += msg + '<br>'
}

pc.ontrack = event => {
  const el = document.createElement(event.track.kind)
  el.srcObject = event.streams[0]
  el.autoplay = true
  el.controls = true
  document.getElementById('remoteVideos').appendChild(el)
}

pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = event => {
  if (event.candidate === null) {
    document.getElementById('localSessionDescription').value = btoa(JSON.stringify(pc.localDescription))
  }
}

// Offer to receive 1 audio, and 1 video tracks
pc.addTransceiver('audio', {
  'direction': 'recvonly'
})
pc.addTransceiver('video', {
  'direction': 'recvonly'
})
pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)

window.startSession = () => {
  const sd = document.getElementById('remoteSessionDescription').value
  if (sd === '') {
    return alert('Session Description must not be empty')
  }
  try {
    pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))))
  } catch (e) {
    alert(e)
  }
}
