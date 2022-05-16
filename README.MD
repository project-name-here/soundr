![Soundr Logo](/resources/logo.svg "Soundr Logo")

# Soundr
Soundr is a simple, open-source, cross-platform audio playing server written in go.
It aims to be simple to setup and work in many envoriments. It is also designed to be
easy to use and maintain. 
Soundr is able to play multiple audio files at the same time. It is able to intigrate well as it uses a REST endpoint.
Swagger Documentation for that endpoint is available in `apiDocs.yml`.
The software it self is written in go and uses the BEEP library. It is made to be shipped as a single executable.
Another target was a minimal dependency tree.
Initally it was written to be used with [Bitfocus Companion](https://bitfocus.io/companion) in a more professional envoriment. (A client for Companion is currently WiP)

# Installation
Installation is as simple as it gets as it is a single executable.
Download one of the releases, drop your sounds into the /sounds folder and run the executable.

# Configuration
The configuration is done in the `conf.json` file.
This is the basic json configuration file layout:
```json
{
    "Port": 8082
}
```

# Usage
Drop your sounds into the /sounds. You can play them by sending a GET request to the /v1/play endpoint.
You need to know the base64 encoded file name of the sound you want to play. You can get started by querying /v1/list. It will return a list of all sounds with their respective base64 encoded file name.
Use that base64 as the `file` parameter in the request.

**Note**: The sounds must be in the format `*.mp3` (more will be supported soon :tm:).

# ToDo
- [ ] Add support for other audio formats