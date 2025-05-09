# MockRPC issue exp

---

~~Just for fun!~~

NOT FUN `X(` !!!!!

## TL;DR

DO NOT VIBE CODING!!!

<https://github.com/pion/webrtc/blob/master/examples/pion-to-pion/offer/main.go>

## What MockRPC is

MockRPC transfers data by WebSocket or WebRTC.

Data format likes jsonrpc but it use `protobuf` by default.

## Bug

A weird bug in this twitter monitor infrastructure:

```plaintext
WebSocket & WebRTC

       <-- ✅ <-- 
CENTER            NODE
       --> ❌ -->

```

- Hard to reproduce, no idea how to reproduce, this repo is used to find out the reason.
- It will only reproduce after reconnecting.
  - The problem cannot be reproduced after reconnecting immediately after startup. You need to run the system normally for a while before reconnecting.
- The connection requires a certain network delay. Our server located at Tokyo and Los Angeles.
  - Center: Tokyo, 2 Core, 1GB mem, Ubuntu 20.04.6 LTS (GNU/Linux 5.15.0-1074-oracle x86_64), go version go1.23.4 linux/amd64, AMD Epyc
  - Node: Los Angeles, 1 Core, 2GB mem, Ubuntu 20.04.6 LTS (GNU/Linux 5.4.0-205-generic x86_64), go version go1.23.4 linux/amd64, Intel maybe?
- more...

## Guess

- maybe here:
  
  ```go
  if len(response) > 0 {
      // <- have an error ?
      if err = rtcContext.Channel.Send(response); err != nil {
          log.Println(rtcContext, response)
      }
  }
  ```

- or

  ```go
  // <- some where I don't known
  ```

## How to test?

```sh
# center
## use pm2
go run main.go --addr=0.0.0.0:11111 --dev=true

# node
## should use public IP address
go run main.go --wsurl=127.0.0.1:11111 --wspwd=node:6:1 --dev=true --ntfy=<THE_KEY>
```

```javascript
// for pm2
/// pm2 reload xxx.js
module.exports = {
    apps: [{
        name: "mockrpc-exp",
        script: "/path/to/center/binary",
        "args": [
            "--addr=0.0.0.0:11111",
            "--dev=true"
        ]
    }]
}
```

If `node` crashes, the reproduction is successful.

## Logs

- 2025-04-10:

  ```txt
  customLogger Warn: Failed to start manager: connecting canceled by caller
  customLogger Warn: Failed to start SCTP: DTLS not established
  customLogger Warn: undeclaredMediaProcessor failed to open SrtpSession: the DTLS transport has not started yet
  customLogger Warn: undeclaredMediaProcessor failed to open SrtcpSession: the DTLS transport has not started yet
  ```

- 2025-04-08:

  first crash cost about `1.6` days, my target is below `12` hrs.

  `1743878429` -> `2025/04/07 09:05:12`
