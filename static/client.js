

function showMessage(str, type) {
    let div = document.createElement("div");
    div.innerHTML = str;
    if(type == "enter") {
        div.style.color = "blue";
    } else if(type == "leave") {
        div.style.color = "red";
    }
    document.body.appendChild(div);
}



var websocket = null
var currentUser = null
var localVideoElement = document.querySelector("video#LocalVideo")
var remoteVideoElement = document.querySelector("video#RemoteVideo")

var localVideoStream = null
var remoteVideoStream = null

var rtcPeerConnection = null

function handleMesg(msg) {
    switch(msg.cmd) {
        case "resp-join": {
            handleJoinRespMsg(msg);
        }
            break;
        case "leave": {
            handleLeaveMsg(msg);
        }
            break;
        case "join": {
            handleRemoteJoinMsg(msg);
        }
            break;
        case "result": {
            handleResultMsg(msg);
        }
            break;
        case "offer": {
            handleRemoteOffer(msg);
        }
            break;
        case "candidate": {
            handleRemoteCandidate(msg);
        }
            break;
        default:
            break;
    }
}

function closeLocalCamera() {
    if(null != localVideoStream) {
        localVideoStream.getVideoTracks()[0].stop()
    }
    localVideoElement.pause();
    localVideoElement.srcObject = null;
    localVideoStream = null
}

function closeRemoteStream() {
    if(null != remoteVideoStream) {
        remoteVideoStream.getVideoTracks()[0].stop()
    }
    remoteVideoElement.pause();
    remoteVideoElement.srcObject = null;
    remoteVideoStream = null
    
    rtcPeerConnection.close()
    rtcPeerConnection = null
}

function openLocalCamera() {
    navigator.mediaDevices.getUserMedia({audio:false, video:true}).then(
        (mediestream)=> {
            localVideoStream = mediestream
            localVideoElement.srcObject = mediestream
            let tracks = mediestream.getTracks()
            tracks[0].addEventListener("ended", (event) => {
                console.log("video stopped....")
            });
        
            tracks[0].onmute = (event)=>{
                console.log("video onmute....")
            }
            console.log(tracks)
        }
    ).catch(
        (error)=> {
            console.log(error)
        }
    )
}

function handleIceCandidate(event) {
    console.info("handleIceCandidate")
    if(event.candidate) {
        let jsonMsg = {
            'cmd':'candidate',
            'roomId':currentUser.roomId,
            'uid':currentUser.userId,
            'msg':JSON.stringify(event.candidate),
        }
        let jsonMsgStr = JSON.stringify(jsonMsg)
        websocket.send(jsonMsgStr)
        console.info("handleIceCandidate info:" + jsonMsgStr)
    } else {
        console.warn("handleIceCandidate fail...")
    }
}

function handleRemoteStreamAdd(event) {
    console.info("handleRemoteStreamAdd")
    remoteVideoStream = event.streams[0]
    remoteVideoElement.srcObject = remoteVideoStream
}

function createOfferAndSendMsg(session) {
    rtcPeerConnection.setLocalDescription(session).then(
        ()=>{
            let jsonMsg = {
                'cmd':'offer',
                'roomId':currentUser.roomId,
                'uid':currentUser.userId,
                'msg':JSON.stringify(session),
            }
            let jsonMsgStr = JSON.stringify(jsonMsg)
            websocket.send(jsonMsgStr)
            console.info("Send sdp info:" + jsonMsgStr)
        }
    ).catch(function() {
        console.warn("setLocalDescription and send fail... ")
    })
}

function handleCreateOfferError() {
    console.warn("handleCreateOffer fail... ")
}

function createPeerConnection() {
    rtcPeerConnection = new RTCPeerConnection()
    rtcPeerConnection.onicecandidate = handleIceCandidate;
    rtcPeerConnection.ontrack = handleRemoteStreamAdd;

    localVideoStream.getTracks().forEach(track=>{
        rtcPeerConnection.addTrack(track, localVideoStream)
    })
}

function openRTCPeerConnection() {
    if(rtcPeerConnection == null) {
        createPeerConnection();
    }
    rtcPeerConnection.createOffer().then(createOfferAndSendMsg).catch(handleCreateOfferError);
}

function handleRemoteOffer(msg) {
    console.info("handleRemoteOffer")
    let shouldCreateOffer = false
    if(rtcPeerConnection == null) {
        createPeerConnection();
        shouldCreateOffer = true;
        
    }
    let desc = JSON.parse(msg.msg)
    rtcPeerConnection.setRemoteDescription(desc)
    if(shouldCreateOffer) {
        rtcPeerConnection.createAnswer().then(createOfferAndSendMsg).catch(handleCreateOfferError);
    }
}
function handleRemoteCandidate(msg) {
    console.info("handleRemoteCandidate", msg)
    var candidate =  JSON.parse(msg.msg)
    rtcPeerConnection.addIceCandidate(candidate).catch(
        (error)=> {
            console.error("addIceCandidate fail..." + e.name)
        }
    )
}

function handleRemoteJoinMsg(msg) {
    showMessage(msg.userName + " join ..." , 'enter')
    setTimeout(() => { // delay 2 seconds 
        openRTCPeerConnection()
    }, 2000)
    
}

function handleJoinRespMsg(msg) {
    currentUser.userId = msg.userId
    showMessage("Join successfully...", 'enter')
    let userNameList = ""
    msg.users.forEach(user => {
        userNameList += user.userName
        userNameList += " ,"
    });
    
    showMessage("Users {" + userNameList +  " } in the room", 'enter')

    openLocalCamera()
}

function handleLeaveMsg(msg) {
    showMessage(msg.userName + " leave..." , 'leave')
}

function handleResultMsg(msg) {
    if(msg.code == 0) {
        showMessage(" Operate success..." , 'enter')
    } else {
        showMessage(" Operate failed..." , 'leave')
    }
}

document.getElementById("connect").onclick = function() {
    websocket = new WebSocket("ws://localhost:5001/websocket")
    websocket.onopen = function() {
        console.log("connect server success...");
        showMessage("connect server success...", 'enter')
        document.getElementById("JoinButton").onclick = function() {
            let roomId = document.getElementById("RoomID").value;
            let userName = document.getElementById("UserName").value;
            currentUser = {
                "roomId": roomId,
                "userName":userName
            }
            let jsonMsg = {
                'cmd':'join',
                'roomId':roomId,
                'userName':userName,
            }

             websocket.send(JSON.stringify(jsonMsg))
        };
        document.getElementById("LeaveButton").onclick = function() {
            let roomId = document.getElementById("RoomID").value;
            let userName = document.getElementById("UserName").value;
            let jsonMsg = {
                'cmd':'leave',
                'userName':currentUser.userName,
                'roomId':roomId,
                'uid':currentUser.userId,
            }

             websocket.send(JSON.stringify(jsonMsg))
             closeLocalCamera()
             closeRemoteStream()
        };
    }

    websocket.onerror = function(env) {
        console.log("Websocket error: " + env.JSON);
        showMessage("Connect fail...", 'leave')
    }

    websocket.onclose = function() {
        console.log("websocket close...")
        showMessage("Connect close...", 'leave')
    }

    websocket.onmessage = function(e) {
        let msg = JSON.parse(e.data);
        handleMesg(msg)
        console.log(e.data)
    }
}


document.getElementById("disconnect").onclick = function() {
    if(websocket) {
        websocket.close();
        document.getElementById("JoinButton").onclick = null;
        document.getElementById("LeaveButton").onclick = null;
    }
    websocket = null
    currentUser = null
}

